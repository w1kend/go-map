package gomap

import (
	"fmt"
	"testing"
)

var sizes = []int{128, 1024, 4096, 8192}

func BenchmarkGet(b *testing.B) {
	for _, n := range sizes {
		mm := New[int64](n)
		for i := 0; i < n; i++ {
			mm.Put(fmt.Sprintf("key__%d", i), int64(i)*2)
		}

		b.Run(fmt.Sprintf("my-map_%d", n), func(b *testing.B) {
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
		mm := New[int64](n)
		b.Run(fmt.Sprintf("my-map_%d", n), func(b *testing.B) {
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
