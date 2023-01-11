package gomap

const (
	bucketSize = 8

	emptyRest       = 0 // this and all other cells with bigger index are empty
	emptyCell       = 1 // there is no value at that index
	evacuatedFirst  = 2 // key/elem is valid.  Entry has been evacuated to first half of larger table.
	evacuatedSecond = 3 // same as above, but evacuated to second half of larger table.
	evacuatedEmpty  = 4 // cell is empty, bucket is evacuated.
	minTopHash      = 5 // minimum topHash value for filled cell
)

// bucket - the Go's bucket explicit representation.
type bucket[K comparable, V any] struct {
	tophash [bucketSize]uint8

	keys   [bucketSize]K
	values [bucketSize]V

	overflow *bucket[K, V]
}

// Get - returns an element for the given key.
// If an element doesn't exist for the given key returns zero value for <V> and false.
func (b *bucket[K, V]) Get(key K, topHash uint8) (V, bool) {
	bkt := b
	for bkt != nil {
		for i := range bkt.tophash {
			top := bkt.tophash[i]
			if top != topHash {
				// if there are no filled cells we break the loop and return zero value
				if top == emptyRest {
					break
				}
				continue
			}

			if !isCellEmpty(top) && bkt.keys[i] == key {
				return bkt.values[i], true
			}
		}

		bkt = bkt.overflow
	}

	return *new(V), false
}

// Put - adds value to the bucket.
// if the value for a given key already exists, it'll be replaced
// if there is no place in this bucket for a new value, new overflow bucket will be created
func (b *bucket[K, V]) Put(key K, topHash uint8, value V) (isAdded bool) {
	var insertIdx int
	var insertBkt *bucket[K, V]

	bkt := b
	for bkt != nil {
		for i := range bkt.tophash {
			// comparing topHash bits, not keys
			// because we can store there flags describing cell state such as cell is empty, cell is evacuating etc.
			// also it's faster than comparing keys
			top := bkt.tophash[i]
			if top != topHash {
				if top == emptyRest {
					insertBkt = bkt
					insertIdx = i
					break
				}

				if insertBkt == nil && isCellEmpty(top) {
					insertBkt = bkt
					insertIdx = i
				}
				continue
			}

			// when we have different keys but tophash is equal
			if bkt.keys[i] != key {
				continue
			}

			bkt.values[i] = value
			return false
		}

		if bkt.overflow == nil {
			// if we didn't find a place to put
			if insertBkt == nil {
				bkt.overflow = &bucket[K, V]{}
				insertBkt = bkt.overflow
				break
			} else { // break if we found a place for the value
				break
			}
		}

		bkt = bkt.overflow
	}

	insertBkt.keys[insertIdx] = key
	insertBkt.values[insertIdx] = value
	insertBkt.tophash[insertIdx] = topHash

	return true
}

func (b *bucket[K, V]) putAt(key K, topHash uint8, value V, idx uint) {
	b.tophash[idx] = topHash
	b.keys[idx] = key
	b.values[idx] = value
}

// Delete - deletes an element with the given key
func (b *bucket[K, V]) Delete(key K, topHash uint8) (deleted bool) {
	bkt := b
	for bkt != nil {
		for i := range bkt.tophash {
			top := bkt.tophash[i]
			if top != topHash {
				// if there are no filled cells we return
				if top == emptyRest {
					return false
				}
				continue
			}

			if bkt.keys[i] == key {
				bkt.tophash[i] = emptyCell
				return true
			}
		}
		bkt = bkt.overflow
	}

	return false
}

func isCellEmpty(val uint8) bool {
	return val <= emptyCell
}

func (b bucket[K, V]) isEvacuated() bool {
	h := b.tophash[0]
	return h > emptyCell && h < minTopHash
}
