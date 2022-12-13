package gomap

const (
	bucketSize = 8

	emptyRest  = 0 // this and all other cells with bigger index are empty
	emptyCell  = 1 // there is no value at that index
	minTopHash = 2 // minimum topHash value for filled cell
)

// bucket - the Go's bucket explicit representation.
type bucket[T any] struct {
	tophash [bucketSize]uint8

	keys   [bucketSize]string
	values [bucketSize]T

	overflow *bucket[T]
}

// Get - returns an element for the given key.
// If an element doesn't exist for the given key returns zero value for <T> and false.
func (b *bucket[T]) Get(key string, topHash uint8) (T, bool) {
	bkt := b
	for bkt != nil {
		for i := range bkt.tophash {
			if bkt.tophash[i] != topHash {
				// if there are no filled cells we break the loop and return zero value
				if bkt.tophash[i] == emptyRest {
					break
				}
				continue
			}

			if !isCellEmpty(bkt.tophash[i]) && bkt.keys[i] == key {
				return bkt.values[i], true
			}
		}

		bkt = bkt.overflow
	}

	return *new(T), false
}

// Put - adds value to the bucket.
// if the value for a given key already exists, it'll be replaced
// if there is no place in this bucket for a new value, new overflow bucket will be created
func (b *bucket[T]) Put(key string, topHash uint8, value T) (isAdded bool) {
	var insertIdx *int

	bkt := b
	for bkt != nil {
		for i := range bkt.tophash {
			// comparing topHash bits, not keys
			// because we can store there flags describing cell state such as cell is empty, cell is evacuating etc.
			// also it's faster than comparing keys
			if bkt.tophash[i] != topHash {
				if bkt.tophash[i] == emptyRest {
					insertIdx = new(int)
					*insertIdx = i
					break
				}

				if insertIdx == nil && isCellEmpty(bkt.tophash[i]) {
					insertIdx = new(int)
					*insertIdx = i
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
			// if current bucket is full
			if insertIdx == nil {
				bkt.overflow = &bucket[T]{}
			} else { // break if we found a place for the value
				break
			}
		}

		bkt = bkt.overflow
	}

	bkt.keys[*insertIdx] = key
	bkt.values[*insertIdx] = value
	bkt.tophash[*insertIdx] = topHash

	return true
}

// Delete - deletes an element with the given key
func (b *bucket[T]) Delete(key string, topHash uint8) (deleted bool) {
	bkt := b
	for bkt != nil {
		for i := range bkt.tophash {
			if bkt.tophash[i] != topHash {
				// if there are no filled cells we return
				if bkt.tophash[i] == emptyRest {
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
