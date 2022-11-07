package gomap

import (
	"math/rand"

	"github.com/dgryski/go-wyhash"
)

const (
	// Maximum average load of a bucket that triggers growth is 6.5.
	// Represent as loadFactorNum/loadFactorDen, to allow integer math.
	loadFactorNum = 13
	loadFactorDen = 2
)

// Hmap - map struct
type Hmap[T any] struct {
	len  int
	b    uint8 // log_2 of # of buckets
	seed uint64

	buckets []Bucket[T]
}

// New - creates a new map for <size> elements
func New[T any](size int) *Hmap[T] {
	h := new(Hmap[T])

	h.seed = generateSeed()

	B := uint8(0)
	for overLoadFactor(size, B) {
		B++
	}
	h.b = B

	h.buckets = make([]Bucket[T], bucketsNum(h.b))

	// fmt.Printf("created map:\nseed - %d\nB - %d\nbuckets count - %d\n", h.seed, h.B, bucketsNum(h.B))

	return h
}

// Get - gets value from the map.
// returns zero value for <T> if there is no value for the given key
func (h Hmap[T]) Get(key string) T {
	tophash, targetBucket := h.locateBucket(key)

	return h.buckets[targetBucket].Get(key, tophash)
}

// Put - puts value into the map
func (h Hmap[T]) Put(key string, value T) {
	tophash, targetBucket := h.locateBucket(key)

	if h.buckets[targetBucket].Put(key, tophash, value) {
		h.len++
	}
}

// Delete - deletes an element from the map
func (h Hmap[T]) Delete(key string) {
	tophash, targetBucket := h.locateBucket(key)
	if deleted := h.buckets[targetBucket].Delete(key, tophash); deleted {
		h.len--
	}
}

func (h Hmap[T]) DescribeBucket(key string) string {
	_, targetBucket := h.locateBucket(key)

	return h.buckets[targetBucket].PrintState()
}

// locateBucket - returns bucket index, where to put/search a value
// and tophash value from hash of the given key
func (h Hmap[T]) locateBucket(key string) (tophash uint8, targetBucket uint64) {
	hash := hash(key, h.seed)
	tophash = topHash(hash)
	mask := bucketMask(h.b)

	// calculete target bucket number, from N available
	// mask represents N-1
	// for N=9  it's 0111
	// for N=16 it's 1111, etc.
	// then, using binary and (hash & mask) we can get up to N different values(index of bucket)
	// where to put/search a value for a given key
	targetBucket = hash & mask
	// fmt.Printf(
	// 	"seed - %d\n hash - %.8b(%d)\n tophash - %.8b(%d)\n mask - %.8b(%d)\n targetbucket - %.8b(%d)\n",
	// 	h.seed,
	// 	hash,
	// 	hash,
	// 	tophash,
	// 	tophash,
	// 	mask,
	// 	mask,
	// 	targetBucket,
	// 	targetBucket,
	// )

	return tophash, targetBucket
}

// returns first 8 bits from the val
func topHash(val uint64) uint8 {
	tophash := uint8(val >> 56)
	if tophash < minTopHash {
		tophash += minTopHash
	}
	return tophash
}

// bucketShift returns 1<<b - actual number of buckets
func bucketsNum(b uint8) uint64 {
	// Masking the shift amount allows overflow checks to be elided.
	return 1 << b
}

// bucketMask returns 1<<b - 1
func bucketMask(b uint8) uint64 {
	return bucketsNum(b) - 1
}

func generateSeed() uint64 {
	return rand.Uint64()
}

// overLoadFactor reports whether count items placed in 1<<B buckets is over loadFactor.
func overLoadFactor(size int, B uint8) bool {
	return size > bucketSize && uint64(size) > loadFactorNum*(bucketsNum(B)/loadFactorDen)
}

// hash - returns hash of the key
func hash(key string, seed uint64) uint64 {
	return wyhash.Hash([]byte(key), seed)
}

// Range - iterates through the map and calls the given func for each key, value.
// if the given func returns false, loop breaks.
func (m Hmap[T]) Range(f func(key string, value T) bool) {
	for i := range m.buckets {
		bucket := &m.buckets[i]
		for bucket != nil {
			for j, th := range bucket.tophash {
				// move to the next bucket when there are no values after index j
				if th == emptyRest {
					continue
				}
				// if there is a value at index j
				if th >= minTopHash {
					if !f(bucket.keys[j], bucket.values[j]) {
						return
					}
				}
			}
			// check overflow buckets
			bucket = bucket.overflow
		}
	}
}

// Len - returns the length of the map
func (m Hmap[T]) Len() int {
	return m.len
}
