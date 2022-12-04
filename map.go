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

	ptrSize = 4 << (^uintptr(0) >> 63) // pointer size
)

// hmap - map struct
type hmap[T any] struct {
	len  int
	b    uint8 // log_2 of # of buckets
	seed uint64

	buckets []Bucket[T]
}

type Hashmap[T any] interface {
	// gets the value for the given key.
	// returns zero value for <T> if there is no value for the given key
	Get(key string) T
	// gets the value for the given key and the flag indicating whether the value exists
	// returns zero value for <T> and false if there is no value for the given key
	Get2(key string) (T, bool)
	// puts value into the map
	Put(key string, value T)
	// deletes an element from the map
	Delete(key string)
	// iterates through the map and calls the given func for each key, value.
	// if the given func returns false, loop breaks.
	Range(f func(k string, v T) bool)
	// returns the length of the map
	Len() int
}

// New - creates a new map for <size> elements
func New[T any](size int) Hashmap[T] {
	h := new(hmap[T])

	h.seed = generateSeed()

	B := uint8(0)
	for overLoadFactor(size, B) {
		B++
	}
	h.b = B

	h.buckets = make([]Bucket[T], bucketsNum(h.b))

	return h
}

func (h hmap[T]) Get(key string) T {
	tophash, targetBucket := h.locateBucket(key)

	v, _ := h.buckets[targetBucket].Get(key, tophash)
	return v
}

func (h hmap[T]) Get2(key string) (T, bool) {
	tophash, targetBucket := h.locateBucket(key)

	return h.buckets[targetBucket].Get(key, tophash)
}

func (h hmap[T]) Put(key string, value T) {
	tophash, targetBucket := h.locateBucket(key)

	if h.buckets[targetBucket].Put(key, tophash, value) {
		h.len++
	}
}

func (h hmap[T]) Delete(key string) {
	tophash, targetBucket := h.locateBucket(key)
	if deleted := h.buckets[targetBucket].Delete(key, tophash); deleted {
		h.len--
	}
}

// locateBucket - returns bucket index, where to put/search a value
// and tophash value from hash of the given key
func (h hmap[T]) locateBucket(key string) (tophash uint8, targetBucket uint64) {
	hash := hash(key, h.seed)
	tophash = topHash(hash)
	mask := bucketMask(h.b)

	// calculate target bucket number, from N available
	// mask represents N-1
	// for N=9  it's 0111
	// for N=16 it's 1111, etc.
	// then, using binary and (hash & mask) we can get up to N different values(index of bucket)
	// where to put/search a value for a given key
	targetBucket = hash & mask

	return tophash, targetBucket
}

// returns first 8 bits from the val
func topHash(val uint64) uint8 {
	tophash := uint8(val >> (ptrSize*8 - 8))
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

func (m hmap[T]) Range(f func(k string, v T) bool) {
	for i := range m.randRangeSequence() {
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

func (m hmap[T]) randRangeSequence() []int {
	i := rand.Intn(len(m.buckets))

	seq := make([]int, 0, len(m.buckets))
	for len(seq) != len(m.buckets) {
		seq = append(seq, i)
		i++
		if i >= len(m.buckets) {
			i = 0
		}
	}

	return seq
}

func (m hmap[T]) Len() int {
	return m.len
}
