package gomap

import (
	"math/rand"

	"github.com/dgryski/go-wyhash"
)

const (
	bucketCnt = 8

	// Maximum average load of a bucket that triggers growth is 6.5.
	// Represent as loadFactorNum/loadFactorDen, to allow integer math.
	loadFactorNum = 13
	loadFactorDen = 2
)

type Hmap[T any] struct {
	Len  int
	B    uint8 // log_2 of # of buckets
	seed uint64

	buckets []Bucket[T]
}

func New[T any](size int) *Hmap[T] {
	h := new(Hmap[T])

	h.seed = generateSeed()

	B := uint8(0)
	for overLoadFactor(size, B) {
		B++
	}
	h.B = B

	h.buckets = make([]Bucket[T], size)

	return h
}

func (h Hmap[T]) Get(key string) T {
	tophash, targetBucket := h.locateBucket(key)

	return h.buckets[targetBucket].Get(key, tophash)
}

func (h Hmap[T]) Put(key string, value T) {
	tophash, targetBucket := h.locateBucket(key)

	h.buckets[targetBucket].Put(key, tophash, value)
}

func (h Hmap[T]) locateBucket(key string) (tophash uint8, targetBucket uint64) {
	hash := hash(key, h.seed)
	tophash = topHash(hash)
	mask := bucketsNum(h.B)
	targetBucket = hash & mask

	return tophash, targetBucket
}

// returns first 8 bits from the val
func topHash(val uint64) uint8 {
	return uint8(val >> 56)
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
	return size > bucketCnt && uint64(size) > loadFactorNum/loadFactorDen*bucketsNum(B)
}

// hash - returns hash of the key
func hash(key string, seed uint64) uint64 {
	return wyhash.Hash([]byte(key), seed)
}
