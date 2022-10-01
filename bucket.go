package gomap

const (
	bucketSize = 8

	zeroIdx = -1
)

type Bucket[T any] struct {
	len     int
	tophash [bucketSize]uint8

	keys   [bucketSize]string
	values [bucketSize]T
}

func (b Bucket[T]) Get(key string, _ uint8) T {
	if i := b.indexOf(key); i != zeroIdx {
		return b.values[i]
	}
	return *new(T)
}

func (b *Bucket[T]) Put(key string, _ uint8, value T) {
	if i := b.indexOf(key); i != zeroIdx {
		b.values[i] = value
		return
	}

	// check bounds
	if b.len > bucketSize {
		return
	}

	b.keys[b.len] = key
	b.values[b.len] = value
	b.len++
}

func (b *Bucket[T]) Delete(key string) {
	if i := b.indexOf(key); i != zeroIdx {
		b.keys[i] = *new(string)
		// if it was the last element
		if i == b.len {
			b.len--
		}
	}
}

func (b Bucket[T]) indexOf(key string) int {
	for i := 0; i < b.len; i++ {
		if b.keys[i] == key {
			return i
		}
	}

	return zeroIdx
}

func (b Bucket[T]) Len() int {
	return b.len
}
