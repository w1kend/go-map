package gomap

import (
	"math/rand"
)

const noCheck uint64 = 1<<(8*ptrSize) - 1

// A hash iteration structure.
type hiter[K comparable, V any] struct {
	key           *K
	elem          *V
	m             *hmap[K, V]
	buckets       *[]bucket[K, V] // bucket ptr at hash_iter initialization time
	currBktPtr    *bucket[K, V]   // current bucket
	startBucket   uint64          // bucket iteration started at
	offset        uint8           // intra-bucket offset to start from during iteration (should be big enough to hold bucketCnt-1)
	wrapped       bool            // already wrapped around from end of bucket array to beginning
	B             uint8
	i             uint8
	currBucketNum uint64
}

func iterInit[K comparable, V any](m *hmap[K, V]) *hiter[K, V] {
	h := hiter[K, V]{}

	if m == nil || m.len == 0 {
		return &h
	}

	h.m = m
	h.B = m.B
	h.buckets = &m.buckets
	h.startBucket = rand.Uint64() & bucketMask(m.B) // pick random bucket
	// choose offset to start from inside a bucket
	h.offset = uint8(uint8(h.startBucket) >> h.B & (bucketSize - 1))
	h.currBucketNum = h.startBucket

	h.m.flags |= iterator | oldIterator // set iterators flags
	h.next()

	return &h
}

func (it *hiter[K, V]) next() {
	b := it.currBktPtr
	bucketNum := it.currBucketNum
	i := it.i
next:
	// choose bucket
	if b == nil {
		if bucketNum == it.startBucket && it.wrapped {
			// end of iteration
			it.key = nil
			it.elem = nil
			return
		}
		b = &it.m.buckets[bucketNum]

		bucketNum++
		if bucketNum == bucketsNum(it.B) {
			bucketNum = 0
			it.wrapped = true
		}
		i = 0
	}

	// iterate over the bucket
	for ; i < bucketSize; i++ {
		// index with offset
		offI := (i + it.offset) & (bucketSize - 1)
		top := b.tophash[offI]
		// we don't check emptyRest as we start iterating in the middle of a bucket
		if isCellEmpty(top) || top == evacuatedEmpty {
			continue
		}
		key := &b.keys[offI]
		elem := &b.values[offI]

		it.key = key
		it.elem = elem

		// update iteration state and return
		it.currBucketNum = bucketNum
		if it.currBktPtr != b {
			it.currBktPtr = b
		}
		it.i = i + 1
		return
	}

	// go to an overflow when finished with the current bucket
	b = b.overflow
	i = 0
	goto next
}
