package gomap

import (
	"fmt"
	"strings"

	"github.com/dolthub/maphash"
)

const (
	// Maximum average load of a bucket that triggers growth is 6.5.
	// Represent as loadFactorNum/loadFactorDen, to allow integer math.
	loadFactorNum = 13
	loadFactorDen = 2

	ptrSize = 4 << (^uintptr(0) >> 63) // pointer size

	// flags
	iterator     = 1 // there may be an iterator using buckets
	oldIterator  = 2 // there may be an iterator using oldbuckets
	hashWriting  = 4 // a goroutine is writing to the map
	sameSizeGrow = 8 // the current map growth is to a new map of the same size
)

// hmap - map struct
type hmap[K comparable, V any] struct {
	len int
	B   uint8 // log_2 of # of buckets

	buckets []bucket[K, V]
	hasher  maphash.Hasher[K] // Go's runtime hasher

	oldbuckets   *[]bucket[K, V]
	numEvacuated uint64 // progress counter for evacuation (buckets less than this have been evacuated)

	flags uint8
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
	String() string
}

// New - creates a new map for <size> elements
func New[K comparable, V any](size int) Hashmap[K, V] {
	h := new(hmap[K, V])

	B := uint8(0)
	for overLoadFactor(size, B) {
		B++
	}
	h.B = B

	h.buckets = make([]bucket[K, V], bucketsNum(h.B))
	h.hasher = maphash.NewHasher[K]()

	return h
}

func (h *hmap[K, V]) Get(key K) V {
	v, _ := h.Get2(key)
	return v
}

func (h *hmap[K, V]) Get2(key K) (V, bool) {
	tophash, targetBucket := h.locateBucket(key)

	b := &h.buckets[targetBucket]

	if h.isGrowing() {
		oldB := &(*h.oldbuckets)[targetBucket&h.oldBucketMask()]
		if !oldB.isEvacuated() {
			b = oldB
		}
	}

	return b.Get(key, tophash)
}

func (h *hmap[K, V]) Put(key K, value V) {
	tophash, targetBucket := h.locateBucket(key)

	// start growing if adding an element will trigger overload
	if !h.isGrowing() && overLoadFactor(h.len+1, h.B) {
		h.startGrowth()
	}

	// evacuate old bucket first
	if h.isGrowing() {
		h.growWork(targetBucket)
	}

	if h.buckets[targetBucket].Put(key, tophash, value) {
		h.len++
	}
}

func (h *hmap[K, V]) Delete(key K) {
	tophash, targetBucket := h.locateBucket(key)

	b := &h.buckets[targetBucket]

	if h.isGrowing() {
		oldB := &(*h.oldbuckets)[targetBucket&h.oldBucketMask()]
		if !oldB.isEvacuated() {
			b = oldB
		}
	}

	if deleted := b.Delete(key, tophash); deleted {
		h.len--
	}
}

// locateBucket - returns bucket index, where to put/search a value
// and tophash value from hash of the given key
func (h *hmap[K, V]) locateBucket(key K) (tophash uint8, targetBucket uint64) {
	hash := h.hasher.Hash(key)
	tophash = topHash(hash)
	mask := bucketMask(h.B)

	// calculate target bucket number, from N available
	// mask represents N-1
	// for N=9  it's 0111
	// for N=16 it's 1111, etc.
	// then, using binary and (hash & mask) we can get up to N different values(index of bucket)
	// where to put/search a value for a given key
	targetBucket = hash & mask

	return tophash, targetBucket
}

func (h *hmap[K, V]) String() string {
	buf := strings.Builder{}
	buf.WriteString("go-map[")
	h.Range(func(k K, v V) bool {
		buf.WriteString(fmt.Sprintf("%v:%v ", k, v))
		return true
	})

	return strings.TrimRight(buf.String(), " ") + "]"
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

func (m *hmap[K, V]) Range(f func(k K, v V) bool) {
	iter := iterInit(m)
	for iter.key != nil && iter.elem != nil {
		if !f(*iter.key, *iter.elem) {
			break
		}
		iter.next()
	}
}

func (m *hmap[K, V]) Len() int {
	return m.len
}

// sameSizeGrow reports whether the current growth is to a map of the same size.
func (h *hmap[K, V]) sameSizeGrow() bool {
	return h.flags&sameSizeGrow != 0
}

func (m *hmap[K, V]) isGrowing() bool {
	return m.oldbuckets != nil
}

func (m *hmap[K, V]) growWork(bucket uint64) {
	// make sure we evacuate the oldbucket corresponding
	// to the bucket we're about to use
	m.evacuate(bucket & m.oldBucketMask())

	// evacuate one more oldbucket to make progress on growing
	if m.isGrowing() {
		m.evacuate(m.numEvacuated)
	}
}

func (m *hmap[K, V]) evacuate(oldbucket uint64) {
	b := &(*m.oldbuckets)[oldbucket]
	newBit := m.numOldBuckets()

	if !b.isEvacuated() {
		// two halfs of the new buckets
		halfs := [2]evacDst[K, V]{{b: &m.buckets[oldbucket]}}

		if !m.sameSizeGrow() {
			// Only calculate y pointers if we're growing bigger.
			// Otherwise GC can see bad pointers.
			halfs[1].b = &m.buckets[oldbucket+newBit]
		}

		for ; b != nil; b = b.overflow {
			// moving all values from the old bucket to the new one
			for i := 0; i < bucketSize; i++ {
				top := b.tophash[i]

				if isCellEmpty(top) {
					b.tophash[i] = evacuatedEmpty
					continue
				}

				key := &b.keys[i]
				value := &b.values[i]

				hash := m.hasher.Hash(*key)

				// decide where to evacuate the element.
				// the first or the second half of the new buckets
				//
				// newBit == # of prev buckets. it's called like that because of it's purpose
				// the value represents new bit of our new mask(# of curr buckets - 1)
				// if newBit == 8 (1000) then newMask == 15(1111) and oldMask == 7(0111)
				// and in that case only the 4th bit(from the end) of mask matters
				// because it decides whether targetBucket changes or not.

				var useSecond uint8
				if !m.sameSizeGrow() && hash&newBit != 0 {
					useSecond = 1
				}

				// evacuatedFirst + useSecond == evaluatedSecond
				b.tophash[i] = evacuatedFirst + useSecond
				dst := &halfs[useSecond]
				// check bounds
				if dst.i == bucketSize {
					dst.b = newOverflow(dst.b)
					dst.i = 0
				}
				dst.b.putAt(*key, top, *value, dst.i)
				dst.i++
			}
		}
	}

	if oldbucket == m.numEvacuated {
		m.advanceEvacuationMark(newBit)
	}
}

func (m *hmap[K, V]) advanceEvacuationMark(newBit uint64) {
	m.numEvacuated++

	stop := newBit + 1024
	if stop > newBit {
		stop = newBit
	}

	for m.numEvacuated != stop && (*m.oldbuckets)[m.numEvacuated].isEvacuated() {
		m.numEvacuated++
	}

	if m.numEvacuated == newBit { // newbit == # of oldbuckets
		// Growing is all done. Free old main bucket array.
		m.oldbuckets = nil
		m.flags &^= sameSizeGrow
	}
}

// evacDst is an evacuation destination.
type evacDst[K comparable, V any] struct {
	b *bucket[K, V] // pointer to the bucket
	i uint          // index for the next element in the destination bucket
}

// noldbuckets calculates the number of buckets prior to the current map growth.
func (m *hmap[K, V]) numOldBuckets() uint64 {
	oldB := m.B
	if !m.sameSizeGrow() {
		oldB--
	}

	return bucketsNum(oldB)
}

// oldbucketmask provides a mask that can be applied to calculate n % noldbuckets().
func (m *hmap[K, V]) oldBucketMask() uint64 {
	return m.numOldBuckets() - 1
}

func (m *hmap[K, V]) startGrowth() {
	oldBuckets := m.buckets
	m.B++
	m.buckets = make([]bucket[K, V], bucketsNum(m.B))
	m.oldbuckets = &oldBuckets
	m.numEvacuated = 0

	flags := m.flags &^ (iterator | oldIterator) // remove iterators flags
	if m.flags&iterator != 0 {
		flags |= oldIterator
	}

	// actual growth happens in the evacuate() and growWork() functions
}

func newOverflow[K comparable, V any](b *bucket[K, V]) *bucket[K, V] {
	if b.overflow == nil {
		b.overflow = &bucket[K, V]{}
	}

	return b.overflow
}

func (m *hmap[K, V]) debug() {
	fmt.Println("main buckets:")
	for i, b := range m.buckets {
		bk := &b
		for bk != nil {
			fmt.Printf("\t\t%d - %s\n", i, bk.debug())
			bk = bk.overflow
		}
	}

	if m.oldbuckets != nil {
		fmt.Println("old buckets:")
		for i, b := range *m.oldbuckets {
			bk := &b
			for bk != nil {
				fmt.Printf("\t\t%d - %s\n", i, bk.debug())
				bk = bk.overflow
			}
		}
	}
}
