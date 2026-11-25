//go:build wasip1 && wasm && !tinygo

package main

import (
	"runtime"
	"slices"

	"tests/maps/foo/foo/maps"

	"github.com/kjanat/go-modules/cm"
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
		wantItems := []string{"nested", "string"}
		lists := maps.VariantMap(maps.EntryTags(cm.ToMap(map[uint32]cm.List[string]{3: cm.ToList(wantItems)})))
		tags := lists.Tags()
		if tags == nil {
			panic("missing variant tags")
		}
		items, found := tags.Get(3)
		if !found || !slices.Equal(items.Slice(), wantItems) {
			panic("bad variant")
		}
		wantWide := [3]uint64{1 << 63, 1 << 40, 23}
		wide := maps.VariantMap(maps.EntryWide(wantWide))
		if v := wide.Wide(); v == nil || *v != wantWide {
			panic("bad wide variant")
		}
		opt := maps.OptionMap(cm.Some(cm.ToMap(map[bool]string{true: "in"})))
		if v, _ := opt.Some().Get(4); v != -3 {
			panic(v)
		}
		errMap := maps.ResultMap()
		if err, _ := errMap.Err().Get(-7); err != "out" {
			panic(err)
		}
		var outcome maps.MixedResult
		outcome.SetErr(cm.ToMap(map[int32]string{-7: "out"}))
		got := maps.MixedResultRoundtrip(outcome)
		errs := got.Err()
		if errs == nil {
			panic("missing result map")
		}
		if err, found := errs.Get(-7); !found || err != "out" {
			panic("bad result map")
		}
		outcome.SetOK(wantWide)
		okResult := maps.MixedResultRoundtrip(outcome)
		if v := okResult.OK(); v == nil || *v != wantWide {
			panic("bad wide result")
		}
		nested := maps.Nested(cm.ToMap(map[string]cm.Map[uint8, float64]{"in": cm.ToMap(map[uint8]float64{3: 6.5})}))
		inner, _ := nested.Get(-5)
		if n, _ := inner.Get("out"); n != 6 {
			panic(n)
		}
		runtime.GC()
		result, found = cfg.Counts.Get("x")
		if !found || result != i+23 {
			panic("bad count after GC")
		}
	}
	return result
}
func main() {}
