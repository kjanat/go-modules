//go:build wasip1 && wasm && !tinygo

package main

import (
	"github.com/kjanat/go-modules/cm"
	"runtime"
	"tests/maps/foo/foo/maps"
)

//go:wasmexport check
func check() uint32 {
	var result uint32
	for i := range uint32(20) {
		x := maps.NamedRoundtrip(maps.Names(cm.ToMap(map[string]uint32{"x": i + 23})))
		cfg := maps.RecordMap(maps.Config{Labels: cm.ToMap(map[string]string{"l": "longer string"}), Counts: x})
		if s, _ := cfg.Labels.Get("l"); s != "longer string" {
			panic(s)
		}
		lists := maps.VariantMap(maps.EntryTags(cm.ToMap(map[uint32]cm.List[string]{3: cm.ToList([]string{"nested", "string"})})))
		items, _ := lists.Tags().Get(3)
		if items.Slice()[1] != "string" {
			panic("bad variant")
		}
		opt := maps.OptionMap(cm.Some(cm.ToMap(map[bool]string{true: "in"})))
		if v, _ := opt.Some().Get(4); v != -3 {
			panic(v)
		}
		errMap := maps.ResultMap()
		if err, _ := errMap.Err().Get(-7); err != "out" {
			panic(err)
		}
		nested := maps.Nested(cm.ToMap(map[string]cm.Map[uint8, float64]{"in": cm.ToMap(map[uint8]float64{3: 6.5})}))
		inner, _ := nested.Get(-5)
		if n, _ := inner.Get("out"); n != 6 {
			panic(n)
		}
		runtime.GC()
		result, _ = cfg.Counts.Get("x")
	}
	return result
}
func main() {}
