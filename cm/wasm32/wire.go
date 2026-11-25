//go:build wasip1 && wasm && !tinygo

// Package wasm32 converts between standard Go values and the memory32
// Canonical ABI. It is used by generated bindings.
package wasm32

import (
	"math"
	"sync"
	"unsafe"

	"github.com/kjanat/go-modules/cm"
)

// Arena keeps pointer-free canonical memory and any allocations referenced
// from it alive together.
type Arena struct {
	buffers [][]byte
}

func (a *Arena) Allocate(size, align uint32) uint32 {
	buf, addr := allocate(size, align)
	a.buffers = append(a.buffers, buf)
	return addr
}

func allocate(size, align uint32) ([]byte, uint32) {
	buf := make([]byte, uint64(max(size, 1))+uint64(align)-1)
	base := uintptr(unsafe.Pointer(unsafe.SliceData(buf)))
	addr := (base + uintptr(align) - 1) &^ (uintptr(align) - 1)
	if addr+uintptr(size) > math.MaxUint32 {
		panic("canonical ABI allocation exceeds memory32")
	}
	return buf, uint32(addr)
}

func Load[T any](addr uint32) T         { return *cm.U32ToPointer[T](addr) }
func Store[T any](addr uint32, value T) { *cm.U32ToPointer[T](addr) = value }

var allocations = struct {
	sync.Mutex
	in  map[uint32][]byte
	out map[uint32]*Arena
}{
	in:  make(map[uint32][]byte),
	out: make(map[uint32]*Arena),
}

var runtimeReady bool

func init() { runtimeReady = true }

// The WASI adapter can request its state while the Go runtime is starting.
// Reserve that memory through the runtime's linear-memory allocator.
//
//go:linkname sbrk runtime.sbrk
func sbrk(size uintptr) unsafe.Pointer

//go:linkname procPin runtime.procPin
func procPin() int

//go:linkname procUnpin runtime.procUnpin
func procUnpin()

//go:wasmimport wasi_snapshot_preview1 adapter_monotonic_clock_set_paused
func setMonotonicClockPaused(paused bool)

//go:wasmexport cabi_realloc
func realloc(addr, oldSize, align, size uint32) uint32 {
	if !runtimeReady {
		ptr := uintptr(sbrk(uintptr(size) + uintptr(align) - 1))
		return uint32((ptr + uintptr(align) - 1) &^ (uintptr(align) - 1))
	}
	// The canonical ABI prohibits leaving the component from realloc.
	// Prevent GC assistance and service monotonic clock reads in the adapter.
	procPin()
	setMonotonicClockPaused(true)
	defer procUnpin()
	defer setMonotonicClockPaused(false)
	allocations.Lock()
	defer allocations.Unlock()
	if size == 0 {
		delete(allocations.in, addr)
		return 0
	}
	buf, next := allocate(size, align)
	if addr != 0 {
		copy(unsafe.Slice(cm.U32ToPointer[byte](next), size),
			unsafe.Slice(cm.U32ToPointer[byte](addr), min(oldSize, size)))
		delete(allocations.in, addr)
	}
	allocations.in[next] = buf
	return next
}

// ReleaseInput drops the host allocation after it has been decoded.
func ReleaseInput(addr uint32) {
	allocations.Lock()
	delete(allocations.in, addr)
	allocations.Unlock()
}

// RetainResult keeps a result alive until the canonical post-return call.
func (a *Arena) RetainResult(addr uint32) {
	allocations.Lock()
	allocations.out[addr] = a
	allocations.Unlock()
}

func ReleaseResult(addr uint32) {
	allocations.Lock()
	delete(allocations.out, addr)
	allocations.Unlock()
}

func LoadString(addr uint32) string {
	data, size := Load[uint32](addr), Load[uint32](addr+4)
	if size == 0 {
		ReleaseInput(data)
		return ""
	}
	value := string(unsafe.Slice(cm.U32ToPointer[byte](data), size))
	ReleaseInput(data)
	return value
}

func StoreString(a *Arena, addr uint32, value string) {
	if len(value) == 0 {
		Store(addr, uint32(0))
		Store(addr+4, uint32(0))
		return
	}
	size := uint32(len(value))
	data := a.Allocate(size, 1)
	copy(unsafe.Slice(cm.U32ToPointer[byte](data), size), value)
	Store(addr, data)
	Store(addr+4, size)
}

func LoadSequence[T any](addr, stride uint32, load func(uint32) T) []T {
	data, count := Load[uint32](addr), Load[uint32](addr+4)
	values := make([]T, count)
	for i := range values {
		values[i] = load(data + uint32(i)*stride)
	}
	ReleaseInput(data)
	return values
}

func StoreSequence[T any](a *Arena, addr, stride, align uint32, values []T, store func(*Arena, uint32, T)) {
	if len(values) == 0 {
		Store(addr, uint32(0))
		Store(addr+4, uint32(0))
		return
	}
	if uint64(len(values))*uint64(stride) > math.MaxUint32 {
		panic("canonical ABI sequence exceeds memory32")
	}
	data := a.Allocate(uint32(len(values))*stride, align)
	for i, value := range values {
		store(a, data+uint32(i)*stride, value)
	}
	Store(addr, data)
	Store(addr+4, uint32(len(values)))
}

func StoreMap[K comparable, V any, M cm.AnyMap[K, V]](a *Arena, addr, stride, align uint32, value M, store func(*Arena, uint32, cm.Tuple[K, V])) {
	data, count := cm.LowerMap[cm.Tuple[K, V]](value)
	StoreSequence(a, addr, stride, align, unsafe.Slice(data, count), store)
}

func LoadMap[M cm.AnyMap[K, V], K comparable, V any](addr, stride uint32, load func(uint32) cm.Tuple[K, V]) M {
	values := LoadSequence(addr, stride, load)
	return cm.LiftMap[M](unsafe.SliceData(values), len(values))
}
