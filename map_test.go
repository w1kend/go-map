package gomap

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestMap(t *testing.T) {
	mm := New[int64](8)

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
		mm := New[int](8)
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
	mm := New[int](8)

	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	prefix := "key_"

	for _, v := range values {
		mm.Put(fmt.Sprintf("%s%d", prefix, v), v)
	}

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
	m := New[NestedStruct](10)

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
		m := New[string](1)
		m.Put(key, key)
		if v := m.Get(key); v != key {
			t.Fatal(v, "!==", key)
		}
	})
}

func TestRange(t *testing.T) {
	m := New[int64](100)

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
