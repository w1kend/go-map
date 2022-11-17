package gomap

import (
	"fmt"
)

const (
	bucketSize = 8

	emptyRest      = 0 // this and all other cells with bigger index are empty
	emptyCell      = 1 // there is no value at that index
	evacuatedX     = 2 // key/elem is valid.  Entry has been evacuated to first half of larger table.
	evacuatedY     = 3 // same as above, but evacuated to second half of larger table.
	evacuatedEmpty = 4 // cell is empty, bucket is evacuated.
	minTopHash     = 5 // minimum topHash value for filled cell
)

type Bucket[T any] struct {
	tophash [bucketSize]uint8

	keys   [bucketSize]string
	values [bucketSize]T

	overflow *Bucket[T]
}

// Get - returns an element for the given key.
// If an element doesn't exist for the given key returns zero value for <T> and false.
func (b Bucket[T]) Get(key string, topHash uint8) (T, bool) {
	for i := range b.tophash {
		if b.tophash[i] != topHash {
			// if there are no filled cells we break the loop and return zero value
			if b.tophash[i] == emptyRest {
				break
			}
			continue
		}

		if !isCellEmpty(b.tophash[i]) && b.keys[i] == key {
			return b.values[i], true
		}
	}

	// check if the key exists in the overflow bucket
	if b.overflow != nil {
		return b.overflow.Get(key, topHash)
	}

	return *new(T), false
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
		return false
	}

	// we have no space in this bucket. check overflow or create a new one
	if insertIdx == nil {
		if b.overflow == nil {
			b.overflow = &Bucket[T]{}
		}

		return b.overflow.Put(key, topHash, value)
	}

	b.keys[*insertIdx] = key
	b.values[*insertIdx] = value
	b.tophash[*insertIdx] = topHash

	return true
}

// Delete - deletes an element with the given key
func (b *Bucket[T]) Delete(key string, topHash uint8) (deleted bool) {
	for i := range b.tophash {
		if b.tophash[i] != topHash {
			// if there are no filled cells we return
			if b.tophash[i] == emptyRest {
				return false
			}
			continue
		}

		if b.keys[i] == key {
			b.tophash[i] = emptyCell
			return true
		}
	}

	// check if the key exists in the overflow bucket
	if b.overflow != nil {
		return b.overflow.Delete(key, topHash)
	}

	return false
}

func isCellEmpty(val uint8) bool {
	return val <= emptyCell
}

func (b Bucket[T]) PrintState() string {
	str := "========================\n"
	str += fmt.Sprintln("tophash", b.tophash)
	str += fmt.Sprintln("keys", b.keys)
	str += fmt.Sprintln("values", b.values)
	str += fmt.Sprintln("overflow", b.overflow)
	str += fmt.Sprintln("========================")

	return str
}
