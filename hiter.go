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
	checkBucket   uint64
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
	checkBucket := it.checkBucket
next:
	// choose bucket
	if b == nil {
		if bucketNum == it.startBucket && it.wrapped {
			// end of iteration
			it.key = nil
			it.elem = nil
			return
		}

		// check old buckets if gwoth is not done
		// skip it if growth started during iteration
		if it.m.isGrowing() && it.B == it.m.B {
			// runtime/map.go:890
			// Iterator was started in the middle of a grow, and the grow isn't done yet.
			// If the bucket we're looking at hasn't been filled in yet (i.e. the old
			// bucket hasn't been evacuated) then we need to iterate through the old
			// bucket and only return the ones that will be migrated to this bucket.
			oldBucketNum := bucketNum & it.m.oldBucketMask()
			b = &(*it.m.oldbuckets)[oldBucketNum]
			if !b.isEvacuated() {
				checkBucket = bucketNum
			} else {
				checkBucket = noCheck
				b = &it.m.buckets[bucketNum]
			}
		} else {
			checkBucket = noCheck
			b = &it.m.buckets[bucketNum]
		}

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

		if checkBucket != noCheck && !it.m.sameSizeGrow() {
			// runtime/map.go:925
			// Special case: iterator was started during a grow to a larger size
			// and the grow is not done yet. We're working on a bucket whose
			// oldbucket has not been evacuated yet. Or at least, it wasn't
			// evacuated when we started the bucket. So we're iterating
			// through the oldbucket, skipping any keys that will go
			// to the other new bucket (each oldbucket expands to two
			// buckets during a grow).

			if key == key {
				hash := it.m.hasher.Hash(*key)
				if hash&bucketMask(it.B) != checkBucket {
					continue
				}
			} else {
				// runtime/map.go:941
				// Hash isn't repeatable if k != k (NaNs).  We need a
				// repeatable and randomish choice of which direction
				// to send NaNs during evacuation. We'll use the low
				// bit of tophash to decide which way NaNs go.
				// NOTE: this case is why we need two evacuate tophash
				// values, evacuatedX and evacuatedY, that differ in
				// their low bit.
				if checkBucket>>(it.B-1) != uint64(b.tophash[offI]&1) {
					continue
				}
			}
		}

		if (top != evacuatedFirst && top != evacuatedSecond) || key != key {
			// This is the golden data, we can return it.
			it.key = key
			it.elem = elem
		} else {
			// The hash table has grown since the iterator was started.
			// The golden data for this key is now somewhere else.
			// Check the current hash table for the data.
			//
			// This code handles the case where the key
			// has been deleted, updated, or deleted and reinserted.
			// NOTE: we need to regrab the key as it has potentially been
			// updated to an equal() but not identical key (e.g. +0.0 vs -0.0).
			re, ok := it.m.Get2(*key) // todo: add getK method
			if !ok {
				continue // key has been deleted
			}
			it.key = key
			it.elem = &re
		}

		// update iteration state and return
		it.currBucketNum = bucketNum
		if it.currBktPtr != b {
			it.currBktPtr = b
		}
		it.i = i + 1
		it.checkBucket = checkBucket
		return
	}

	// go to an overflow when finished with the current bucket
	b = b.overflow
	i = 0
	goto next
}
