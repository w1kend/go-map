package gomap

const (
	bucketSize = 8

	zeroIdx = -1
)

type Bucket[T any] struct {
	k []string
	v []T
}

func NewBucket[T any]() *Bucket[T] {
	return &Bucket[T]{
		k: make([]string, 0, bucketSize),
		v: make([]T, 0, bucketSize),
	}
}

func (b Bucket[T]) Get(key string) T {
	if i := b.indexOf(key); i != zeroIdx {
		return b.v[i]
	}
	return *new(T)
}

func (b *Bucket[T]) Put(key string, value T) {
	if i := b.indexOf(key); i != zeroIdx {
		b.v[i] = value
		return
	}

	b.k = append(b.k, key)
	b.v = append(b.v, value)
}

func (b Bucket[T]) indexOf(key string) int {
	for i := range b.k {
		if b.k[i] == key {
			return i
		}
	}

	return zeroIdx
}

func (b Bucket[T]) Len() int {
	return len(b.k)
}
