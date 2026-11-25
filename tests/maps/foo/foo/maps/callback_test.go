//go:build !wasip1 || tinygo

package maps

import (
	"testing"
	"unsafe"

	"github.com/kjanat/go-modules/cm"
)

func TestNamedRoundtripDuplicateKeys(t *testing.T) {
	pairs := []cm.Tuple[string, uint32]{{F0: "x", F1: 1}, {F0: "x", F1: 2}}
	called := false
	Exports.NamedRoundtrip = func(x Names) Names {
		called = true
		if v, ok := x.Get("x"); !ok || v != 2 {
			t.Errorf("Get(x) = %d, %t; want 2, true", v, ok)
		}
		n := 0
		for k, v := range x.All() {
			n++
			if k != "x" || v != 2 {
				t.Errorf("All() yielded %q, %d; want x, 2", k, v)
			}
		}
		if n != 1 {
			t.Errorf("All() yielded %d entries; want 1", n)
		}
		if l := x.Len(); l != 1 {
			t.Errorf("Len() = %d; want 1", l)
		}
		return x
	}
	result := wasmexport_NamedRoundtrip(unsafe.SliceData(pairs), uint32(len(pairs)))
	if !called {
		t.Fatal("wasmexport_NamedRoundtrip did not call Exports.NamedRoundtrip")
	}
	data, n := cm.LowerMap[cm.Tuple[string, uint32]](*result)
	if data != unsafe.SliceData(pairs) || n != 2 {
		t.Errorf("lowered result = %p, %d; want %p, 2", data, n, unsafe.SliceData(pairs))
	}
}
