package cm

import (
	"slices"
	"unsafe"
)

// Map represents a WIT [map] type. Lookups resolve duplicate keys to the last value.
//
// [map]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#types
type Map[K comparable, V any] struct {
	_ HostLayout
	pairs[K, V]
}

type AnyMap[K comparable, V any] interface {
	~struct {
		_ HostLayout
		pairs[K, V]
	}
}

type pairs[K comparable, V any] struct {
	_    HostLayout
	data *Tuple[K, V]
	len  uintptr
}

func ToMap[M ~map[K]V, K comparable, V any](m M) Map[K, V] {
	s := make([]Tuple[K, V], 0, len(m))
	for k, v := range m {
		s = append(s, Tuple[K, V]{F0: k, F1: v})
	}
	return Map[K, V]{data: unsafe.SliceData(s), len: uintptr(len(s))}
}

func (p pairs[K, V]) entries() []Tuple[K, V] {
	return unsafe.Slice(p.data, p.len)
}

func (p pairs[K, V]) Get(key K) (V, bool) {
	for _, e := range slices.Backward(p.entries()) {
		if e.F0 == key {
			return e.F1, true
		}
	}
	var zero V
	return zero, false
}

func (p pairs[K, V]) All() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		s := p.entries()
		last := make(map[K]int, len(s))
		for i, e := range s {
			last[e.F0] = i
		}
		for i, e := range s {
			if last[e.F0] == i && !yield(e.F0, e.F1) {
				return
			}
		}
	}
}

func (p pairs[K, V]) Len() int {
	s := p.entries()
	keys := make(map[K]struct{}, len(s))
	for _, e := range s {
		keys[e.F0] = struct{}{}
	}
	return len(keys)
}

func (p pairs[K, V]) Map() map[K]V {
	s := p.entries()
	m := make(map[K]V, len(s))
	for _, e := range s {
		m[e.F0] = e.F1
	}
	return m
}

func LowerMap[E Tuple[K, V] | [2]K, M AnyMap[K, V], K comparable, V any](m M) (*E, uint32) {
	checkEntry[E, K, V]()
	p := (*Map[K, V])(unsafe.Pointer(&m))
	return (*E)(unsafe.Pointer(p.data)), uint32(p.len)
}

func LiftMap[M AnyMap[K, V], K comparable, V any, E Tuple[K, V] | [2]K, Len AnyInteger](data *E, len Len) M {
	checkEntry[E, K, V]()
	return M(Map[K, V]{data: (*Tuple[K, V])(unsafe.Pointer(data)), len: uintptr(len)})
}

func checkEntry[E Tuple[K, V] | [2]K, K comparable, V any]() {
	if _, ok := any((*E)(nil)).(*[2]K); !ok {
		return
	}
	if _, ok := any((*K)(nil)).(*V); !ok {
		panic("cm: map entry [2]K requires identical key and value types")
	}
}
