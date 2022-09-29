package gomap

import (
	"math/rand"
)

const (
	bucketCnt = 8

	// Maximum average load of a bucket that triggers growth is 6.5.
	// Represent as loadFactorNum/loadFactorDen, to allow integer math.
	loadFactorNum = 13
	loadFactorDen = 2
)

type Hmap struct {
	B     uint8 // log_2 of # of buckets
	hash0 uint32
}

func NewHmap(size int) *Hmap {
	h := new(Hmap)

	h.hash0 = fastrand()

	B := uint8(0)
	for overLoadFactor(size, B) {
		B++
	}
	h.B = B

	return h
}

// bucketShift returns 1<<b
func bucketShift(b uint8) uint {
	// Masking the shift amount allows overflow checks to be elided.
	return 1 << b
}

// bucketMask returns 1<<b - 1
func bucketMask(b uint8) uint {
	return bucketShift(b) - 1
}

func fastrand() uint32 {
	return rand.Uint32()
}

// overLoadFactor reports whether count items placed in 1<<B buckets is over loadFactor.
func overLoadFactor(count int, B uint8) bool {
	return count > bucketCnt && uint(count) > loadFactorNum*(bucketShift(B)/loadFactorDen)
}
