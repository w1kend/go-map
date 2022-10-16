package gomap

import "fmt"

const (
	bucketSize = 8

	// this and all other cells with bigger index are empty
	emptyRest = 0
	// there is no value at that index
	emptyCell = 1
	// minimum topHash value for filled cell
	minTopHash = 5
)

type Bucket[T any] struct {
	len     int
	tophash [bucketSize]uint8

	keys   [bucketSize]string
	values [bucketSize]T

	overflow *Bucket[T]
}

func (b Bucket[T]) Get(key string, topHash uint8) T {
	for i := range b.tophash {
		if b.tophash[i] != topHash {
			// if there are no filled cells we break the loop and return zero value
			if b.tophash[i] == emptyRest {
				break
			}
			continue
		}

		if !isCellEmpty(b.tophash[i]) && b.keys[i] == key {
			return b.values[i]
		}
	}

	// check if the key exists in the overflow bucket
	if b.overflow != nil {
		return b.overflow.Get(key, topHash)
	}

	return *new(T)
}

// Put - adds value to the bucket.
// if the value for a given key already exists, it'll be replaced
// if there is no place in this bucket for a new value, new overflow bucket will be created
func (b *Bucket[T]) Put(key string, topHash uint8, value T) (isAdded bool) {
	var insertIdx *int

	for i := range b.tophash {
		// comparing topHash bits, not keys
		// because we can store there flags describing cell state such as cell is empty, cell is evacuating etc.
		// also it's faster than comparing keys
		if b.tophash[i] != topHash {
			if b.tophash[i] == emptyRest {
				insertIdx = new(int)
				*insertIdx = i
				break
			}

			if insertIdx == nil && isCellEmpty(b.tophash[i]) {
				insertIdx = new(int)
				*insertIdx = i
			}
			continue
		}

		// when we have different keys but tophash is equal
		if b.keys[i] != key {
			continue
		}

		b.values[i] = value
		return
	}

	// we have no space in this bucket. check overflow or create a new one
	if insertIdx == nil {
		if b.overflow == nil {
			b.overflow = &Bucket[T]{}
		}

		isAdded = b.overflow.Put(key, topHash, value)
		return
	}

	b.keys[*insertIdx] = key
	b.values[*insertIdx] = value
	b.tophash[*insertIdx] = topHash
	b.len++
	isAdded = true

	return isAdded
}

// func (b *Bucket[T]) Delete(key string) {
// 	if i := b.indexOf(key); i != zeroIdx {
// 		b.keys[i] = *new(string)
// 		// if it was the last element
// 		if i == b.len {
// 			b.len--
// 		}
// 	}
// }

// Len - returns number of stored elements, including overflow buckets.
func (b Bucket[T]) Len() int {
	l := b.len
	if b.overflow != nil {
		l += b.overflow.Len()
	}

	return l
}

func isCellEmpty(val uint8) bool {
	return val <= emptyCell
}

func (b Bucket[T]) PrintState() string {
	str := "========================\n"
	str += fmt.Sprintln("len", b.len)
	str += fmt.Sprintln("tophash", b.tophash)
	str += fmt.Sprintln("keys", b.keys)
	str += fmt.Sprintln("values", b.values)
	str += fmt.Sprintln("overflow", b.overflow)
	str += fmt.Sprintln("========================")

	return str
}
