package gomap

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestMap(t *testing.T) {
	mm := New[string, int64](8)

	v, ok := mm.Get2("123")
	isEqual(t, ok, false)
	isEqual(t, v, *new(int64))

	mm.Put("key1", 10)
	v = mm.Get("key1")
	isEqual(t, v, int64(10))

	mm.Put("", 144)
	isEqual(t, mm.Get(""), int64(144))

	mm.Put(" ", 145)
	isEqual(t, mm.Get(" "), int64(145))

	mm.Delete("123")
	v, ok = mm.Get2("123")
	isEqual(t, ok, false)
	isEqual(t, v, *new(int64))

	mm.Put("key1", 20)
	v = mm.Get("key1")
	isEqual(t, v, int64(20))

	t.Run("target value in overflow bucket", func(t *testing.T) {
		mm := New[string, int](8)
		mm.Put("key0", 20)

		for i := 0; i < 8; i++ {
			mm.Put(fmt.Sprintf("key_%d", i), i*10)
		}

		mm.Put("key__1", 10)
		// remove space for an element in the bucket
		mm.Delete("key0")

		// try to add a value in a hole. "key__1" now is stored in an overflow bucket
		mm.Put("key__1", 20)
		v := mm.Get("key__1")
		isEqual(t, v, 20)
		// the values must be deleted from the overflow bucket
		mm.Delete("key__1")

		v = mm.Get("key__1")
		isEqual(t, v, 0)
	})
}

func isEqual(t *testing.T, got interface{}, want interface{}) {
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result is not equal\ngot:  %+v\nwant: %+v\n", got, want)
	}
}

func TestBucketOverflow(t *testing.T) {
	// create map with 8 elements(1 bucket)
	mm := New[string, int](8)

	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	prefix := "key_"

	for _, v := range values {
		mm.Put(fmt.Sprintf("%s%d", prefix, v), v)
	}

	dm := mm.(*hmap[string, int])
	dm.debug()

	for _, v := range values {
		got := mm.Get(fmt.Sprintf("%s%d", prefix, v))
		isEqual(t, got, v)
	}
}

type NestedStruct struct {
	A int64
	B struct {
		C string
		D string
		E struct {
			F []int64
		}
	}
}

func TestGet2(t *testing.T) {
	m := New[string, NestedStruct](10)

	emptyStruct := NestedStruct{}
	m.Put("123", emptyStruct)
	got, ok := m.Get2("123")
	isEqual(t, ok, true)
	isEqual(t, got, emptyStruct)

	got, ok = m.Get2("random_key")
	isEqual(t, ok, false)
	isEqual(t, got, emptyStruct)
}

func FuzzMap(f *testing.F) {
	f.Fuzz(func(t *testing.T, key string) {
		m := New[string, string](1)
		m.Put(key, key)
		if v := m.Get(key); v != key {
			t.Fatal(v, "!==", key)
		}
	})
}

func TestRange(t *testing.T) {
	m := New[string, int64](100)

	n := 100
	wantKeys := make([]string, 0, n)
	wantValues := make([]int64, 0, n)

	for i := 0; i < n; i++ {
		k := fmt.Sprintf("k%d", i)
		v := int64(i) * 10
		m.Put(k, v)
		wantKeys = append(wantKeys, k)
		wantValues = append(wantValues, v)
	}

	gotKeys := make([]string, 0, n)
	gotValues := make([]int64, 0, n)
	m.Range(func(k string, v int64) bool {
		gotKeys = append(gotKeys, k)
		gotValues = append(gotValues, v)
		return true
	})

	sort.Strings(wantKeys)
	sort.Strings(gotKeys)
	isEqual(t, gotKeys, wantKeys)

	i64Less := func(s []int64) func(i, j int) bool {
		return func(i, j int) bool {
			return s[i] < s[j]
		}
	}
	sort.Slice(wantValues, i64Less(wantValues))
	sort.Slice(gotValues, i64Less(gotValues))
	isEqual(t, wantValues, gotValues)
}

type testcase[K comparable, V any] struct{}

func (tt testcase[K, V]) test(t *testing.T, keys []K, values []V) {
	if len(keys) != len(values) {
		t.Fatalf("lengths of keys(%d) and values(%d) must be equal", len(keys), len(values))
	}
	m := New[K, V](len(keys))

	for i, k := range keys {
		m.Put(k, values[i])

		got, ok := m.Get2(k)
		isEqual(t, ok, true)
		isEqual(t, got, values[i])
	}
}
func TestDifferentKeyTypes(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		type keyStruct struct {
			key     string
			anyData [1]int
		}

		tests := testcase[keyStruct, string]{}
		tests.test(
			t,
			[]keyStruct{
				{key: "k1"}, {key: "k2", anyData: [1]int{1}}, {key: "k3"}, {key: "k4", anyData: [1]int{2}},
				{key: "k5"}, {key: "k6", anyData: [1]int{3}}, {key: "k7"}, {key: "k8", anyData: [1]int{4}},
				{key: "k9"},
			},
			[]string{"val1", "val2", "val3", "val4", "val5", "val6", "val7", "val8", "val9"},
		)
	})

	t.Run("array", func(t *testing.T) {
		type keyArray [2]int

		tests := testcase[keyArray, string]{}

		tests.test(
			t,
			[]keyArray{{1, 2}, {2, 3}, {3, 4}, {4, 5}, {5, 6}, {6, 7}, {7, 8}, {8, 9}, {9, 10}},
			[]string{"val1", "val2", "val3", "val4", "val5", "val6", "val7", "val8", "val9"},
		)
	})

	t.Run("bools", func(t *testing.T) {
		tests := testcase[bool, int]{}
		tests.test(t, []bool{true, true, false, false, true}, []int{1, 0, 1, 0, 20})
	})

	t.Run("numbers", func(t *testing.T) {
		t.Run("float64", func(t *testing.T) {
			tests := testcase[float64, int]{}
			tests.test(t, []float64{1.1, 2.2, 3.3, 4.4}, []int{1, 2, 3, 4})
		})
		t.Run("uint64", func(t *testing.T) {
			tests := testcase[uint64, int]{}
			tests.test(t, []uint64{1, 2, 3, 4}, []int{1, 2, 3, 4})
		})
		t.Run("int64", func(t *testing.T) {
			tests := testcase[int64, int]{}
			tests.test(t, []int64{1, 2, 3, 4}, []int{1, 2, 3, 4})
		})
		t.Run("complex64", func(t *testing.T) {
			tests := testcase[complex64, int]{}
			tests.test(t, []complex64{1 + 1i, 2 + 2i, 3 + 3i, 4 + 4i}, []int{1, 2, 3, 4})
		})
	})

	t.Run("byte", func(t *testing.T) {
		tests := testcase[byte, int]{}
		tests.test(t, []byte{'1', '2', '3', '4'}, []int{1, 2, 3, 4})
	})

	t.Run("channel", func(t *testing.T) {
		tests := testcase[chan int, int]{}
		ch1, ch2, ch3, ch4 := make(chan int, 1), make(chan int, 1), make(chan int, 1), make(chan int, 1)
		ch2 <- 2
		ch3 <- 3
		ch4 <- 4
		tests.test(t, []chan int{ch1, ch2, ch3, ch4}, []int{1, 2, 3, 4})
	})
}
