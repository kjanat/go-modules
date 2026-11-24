//go:build wasip1 && wasm && !tinygo

package main

import (
	"github.com/kjanat/go-modules/cm"
	"tests/maps/foo/foo/maps"
)

func init() {
	maps.Exports.NamedRoundtrip = func(x maps.Names) maps.Names {
		n, ok := x.Get("x")
		if !ok {
			panic("missing x")
		}
		return maps.Names(cm.ToMap(map[string]uint32{"x": n}))
	}
	maps.Exports.RecordMap = func(x maps.Config) maps.Config { return x }
	maps.Exports.VariantMap = func(x maps.Entry) maps.Entry { return x }
	maps.Exports.OptionMap = func(x cm.Option[cm.Map[bool, string]]) cm.Option[cm.Map[uint16, int8]] {
		if v := x.Some(); v != nil {
			if s, _ := v.Get(true); s == "in" {
				return cm.Some(cm.ToMap(map[uint16]int8{4: -3}))
			}
		}
		return cm.None[cm.Map[uint16, int8]]()
	}
	maps.Exports.ResultMap = func() cm.Result[cm.Map[int16, uint8], cm.Map[int16, uint8], cm.Map[int32, string]] {
		return cm.Err[cm.Result[cm.Map[int16, uint8], cm.Map[int16, uint8], cm.Map[int32, string]]](cm.ToMap(map[int32]string{-7: "out"}))
	}
	maps.Exports.Nested = func(x cm.Map[string, cm.Map[uint8, float64]]) cm.Map[int64, maps.Names] {
		inner, _ := x.Get("in")
		n, _ := inner.Get(3)
		return cm.ToMap(map[int64]maps.Names{-5: maps.Names(cm.ToMap(map[string]uint32{"out": uint32(n)}))})
	}
	maps.Exports.SameRoundtrip = func(x maps.Same) maps.Same { return x }
	maps.Exports.AnonymousRet = func() cm.Map[rune, bool] { return cm.ToMap(map[rune]bool{'z': true}) }
	maps.Exports.AnonymousParam = func(x cm.Map[string, uint64]) {
		if v, _ := x.Get("key"); v != 100 {
			panic(v)
		}
	}
}
func main() {}
