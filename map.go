package gomap

import (
	"math/rand"

	"github.com/dolthub/maphash"
)

const (
	// Maximum average load of a bucket that triggers growth is 6.5.
	// Represent as loadFactorNum/loadFactorDen, to allow integer math.
	loadFactorNum = 13
	loadFactorDen = 2

	ptrSize = 4 << (^uintptr(0) >> 63) // pointer size
)

// hmap - map struct
type hmap[K comparable, V any] struct {
	len int
	b   uint8 // log_2 of # of buckets

	buckets []bucket[K, V]
	hasher  maphash.Hasher[K] // Go's runtime hasher
}

type Hashmap[K comparable, V any] interface {
	// gets the value for the given key.
	// returns zero value for <V> if there is no value for the given key
	Get(key K) V
	// gets the value for the given key and the flag indicating whether the value exists
	// returns zero value for <V> and false if there is no value for the given key
	Get2(key K) (V, bool)
	// puts value into the map
	Put(key K, value V)
	// deletes an element from the map
	Delete(key K)
	// iterates through the map and calls the given func for each key, value.
	// if the given func returns false, loop breaks.
	Range(f func(k K, v V) bool)
	// returns the length of the map
	Len() int
}

// New - creates a new map for <size> elements
func New[K comparable, V any](size int) Hashmap[K, V] {
	h := new(hmap[K, V])

	B := uint8(0)
	for overLoadFactor(size, B) {
		B++
	}
	h.b = B

	h.buckets = make([]bucket[K, V], bucketsNum(h.b))
	h.hasher = maphash.NewHasher[K]()

	return h
}

func (h hmap[K, V]) Get(key K) V {
	tophash, targetBucket := h.locateBucket(key)

	v, _ := h.buckets[targetBucket].Get(key, tophash)
	return v
}

func (h hmap[K, V]) Get2(key K) (V, bool) {
	tophash, targetBucket := h.locateBucket(key)

	return h.buckets[targetBucket].Get(key, tophash)
}

func (h hmap[K, V]) Put(key K, value V) {
	tophash, targetBucket := h.locateBucket(key)

	if h.buckets[targetBucket].Put(key, tophash, value) {
		h.len++
	}
}

func (h hmap[K, V]) Delete(key K) {
	tophash, targetBucket := h.locateBucket(key)
	if deleted := h.buckets[targetBucket].Delete(key, tophash); deleted {
		h.len--
	}
}

// locateBucket - returns bucket index, where to put/search a value
// and tophash value from hash of the given key
func (h hmap[K, V]) locateBucket(key K) (tophash uint8, targetBucket uint64) {
	hash := h.hasher.Hash(key)
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

// overLoadFactor reports whether count items placed in 1<<B buckets is over loadFactor.
func overLoadFactor(size int, B uint8) bool {
	return size > bucketSize && uint64(size) > loadFactorNum*(bucketsNum(B)/loadFactorDen)
}

func (m hmap[K, V]) Range(f func(k K, v V) bool) {
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

func (m hmap[K, V]) randRangeSequence() []int {
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

func (m hmap[K, V]) Len() int {
	return m.len
}
