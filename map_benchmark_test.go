package gomap

import (
	"fmt"
	"testing"
)

var sizes = []int{128, 1024, 8192}

func BenchmarkGet(b *testing.B) {
	for _, n := range sizes {
		mm := New[string, int64](n)
		for i := 0; i < n; i++ {
			mm.Put(fmt.Sprintf("key__%d", i), int64(i)*2)
		}

		b.Run(fmt.Sprintf("generic-map_%d", n), func(b *testing.B) {
			var got int64
			j := 0
			for i := 0; i < b.N; i++ {
				if j > n {
					j = 0
				}
				got = mm.Get(fmt.Sprintf("key__%d", j))
				j++
			}
			_ = got
		})
	}

	for _, n := range sizes {
		stdm := make(map[string]int64, n)
		for i := 0; i < n; i++ {
			stdm[fmt.Sprintf("key__%d", i)] = int64(i) * 2
		}

		b.Run(fmt.Sprintf("std-map_%d", n), func(b *testing.B) {
			var got int64
			j := 0
			for i := 0; i < b.N; i++ {
				if j > n {
					j = 0
				}
				got = stdm[fmt.Sprintf("key__%d", j)]
				j++
			}
			_ = got
		})
	}
}

func BenchmarkPut(b *testing.B) {
	for _, n := range sizes {
		mm := New[string, int64](n)
		b.Run(fmt.Sprintf("generic-map_%d", n), func(b *testing.B) {
			j := 0
			for i := 0; i < b.N; i++ {
				if j > n {
					j = 0
				}
				mm.Put(fmt.Sprintf("key__%d", j), int64(j))
				j++
			}
		})
	}

	for _, n := range sizes {
		stdm := make(map[string]int64, n)
		b.Run(fmt.Sprintf("std-map_%d", n), func(b *testing.B) {
			j := 0
			for i := 0; i < b.N; i++ {
				if j > n {
					j = 0
				}
				stdm[fmt.Sprintf("key__%d", j)] = int64(j)
				j++
			}
		})
	}
}

func BenchmarkPutWithOverflow(b *testing.B) {
	startSize := 1_000
	targetSize := []int{10_000, 100_000, 1_000_000}

	for _, n := range targetSize {
		mm := New[string, int64](startSize)
		b.Run(fmt.Sprintf("generic-map-with-overflow__%d", n), func(b *testing.B) {
			j := 0
			for i := 0; i < b.N; i++ {
				if j > n {
					j = 0
				}
				mm.Put(fmt.Sprintf("key__%d", j), int64(j))
				j++
			}
		})
	}

	for _, n := range targetSize {
		stdm := make(map[string]int64, startSize)
		b.Run(fmt.Sprintf("std-map-with-evacuation__%d", n), func(b *testing.B) {
			j := 0
			for i := 0; i < b.N; i++ {
				if j > n {
					j = 0
				}
				stdm[fmt.Sprintf("key__%d", j)] = int64(j)
				j++
			}
		})
	}
}
