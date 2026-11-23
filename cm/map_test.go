package cm

import (
	"encoding/json"
	"maps"
	"testing"
	"unsafe"
)

func duplicateKeys() Map[string, uint32] {
	s := []Tuple[string, uint32]{{F0: "x", F1: 1}, {F0: "y", F1: 3}, {F0: "x", F1: 2}}
	return LiftMap[Map[string, uint32]](unsafe.SliceData(s), len(s))
}

func TestMapLastValueWins(t *testing.T) {
	m := duplicateKeys()
	if v, ok := m.Get("x"); !ok || v != 2 {
		t.Errorf("Get(x) = %d, %t; want 2, true", v, ok)
	}
	if v, ok := m.Get("z"); ok || v != 0 {
		t.Errorf("Get(z) = %d, %t; want 0, false", v, ok)
	}
	if n := m.Len(); n != 2 {
		t.Errorf("Len() = %d; want 2", n)
	}
	var keys []string
	var values []uint32
	for k, v := range m.All() {
		keys = append(keys, k)
		values = append(values, v)
	}
	if len(keys) != 2 || keys[0] != "y" || values[0] != 3 || keys[1] != "x" || values[1] != 2 {
		t.Errorf("All() yielded %v %v; want [y x] [3 2]", keys, values)
	}
	if got, want := m.Map(), map[string]uint32{"x": 2, "y": 3}; !maps.Equal(got, want) {
		t.Errorf("Map() = %v; want %v", got, want)
	}
}

func TestMapAllStops(t *testing.T) {
	n := 0
	for range duplicateKeys().All() {
		n++
		break
	}
	if n != 1 {
		t.Errorf("All() yielded %d entries after break; want 1", n)
	}
}

func TestLowerMapKeepsPairs(t *testing.T) {
	type names Map[string, uint32]
	m := names(duplicateKeys())
	data, n := LowerMap[Tuple[string, uint32]](m)
	if n != 3 {
		t.Errorf("LowerMap length = %d; want 3", n)
	}
	if got := unsafe.Slice(data, n)[2]; got.F0 != "x" || got.F1 != 2 {
		t.Errorf("LowerMap entry 2 = %v; want {x 2}", got)
	}
}

func TestLiftMapArrayEntries(t *testing.T) {
	s := [][2]string{{"a", "1"}, {"a", "2"}}
	m := LiftMap[Map[string, string]](unsafe.SliceData(s), len(s))
	if v, _ := m.Get("a"); v != "2" {
		t.Errorf("Get(a) = %q; want 2", v)
	}
	data, n := LowerMap[[2]string](m)
	if data != unsafe.SliceData(s) || n != 2 {
		t.Errorf("LowerMap = %p, %d; want %p, 2", data, n, unsafe.SliceData(s))
	}
}

func TestMapArrayEntryNeedsMatchingTypes(t *testing.T) {
	mustPanic(t, "LowerMap", func() { LowerMap[[2]string](Map[string, uint32]{}) })
	mustPanic(t, "LiftMap", func() { LiftMap[Map[string, uint32]]((*[2]string)(nil), 0) })
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s with a [2]K entry for differing key and value types did not panic", name)
		}
	}()
	f()
}

func TestToMap(t *testing.T) {
	want := map[rune]bool{'a': true, 'b': false}
	m := ToMap(want)
	if got := m.Map(); !maps.Equal(got, want) {
		t.Errorf("ToMap(%v).Map() = %v", want, got)
	}
	if _, n := LowerMap[Tuple[rune, bool]](m); n != 2 {
		t.Errorf("ToMap length = %d; want 2", n)
	}
}

func TestMapLayout(t *testing.T) {
	if unsafe.Sizeof(Map[string, uint32]{}) != unsafe.Sizeof(List[Tuple[string, uint32]]{}) {
		t.Error("Map and List of Tuple differ in size")
	}
	if unsafe.Alignof(Map[string, uint32]{}) != unsafe.Alignof(List[Tuple[string, uint32]]{}) {
		t.Error("Map and List of Tuple differ in alignment")
	}
}

func TestMapJSON(t *testing.T) {
	got, err := json.Marshal(duplicateKeys())
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"y":3,"x":2}`; string(got) != want {
		t.Errorf("Marshal = %s; want %s", got, want)
	}

	var m Map[string, uint32]
	if err := json.Unmarshal([]byte(`{"x":1,"x":2}`), &m); err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Get("x"); v != 2 || m.Len() != 1 {
		t.Errorf("Unmarshal duplicate keys: Get(x) = %d, Len() = %d; want 2, 1", v, m.Len())
	}
	if _, n := LowerMap[Tuple[string, uint32]](m); n != 2 {
		t.Errorf("Unmarshal kept %d pairs; want 2", n)
	}
}

func TestMapJSONKeys(t *testing.T) {
	roundTrip(t, ToMap(map[bool]string{true: "t"}), `{"true":"t"}`)
	roundTrip(t, ToMap(map[int64]uint8{-7: 1}), `{"-7":1}`)
	roundTrip(t, ToMap(map[rune]rune{'a': 'b'}), `{"97":98}`)
	roundTrip(t, ToMap(map[string]List[uint8]{"1": ToList([]uint8{1, 2})}), `{"1":[1,2]}`)
	roundTrip(t, Map[uint16, bool]{}, `{}`)
}

func roundTrip[K comparable, V any](t *testing.T, m Map[K, V], want string) {
	t.Helper()
	got, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("Marshal = %s; want %s", got, want)
	}
	var back Map[K, V]
	if err := json.Unmarshal(got, &back); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(back)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != want {
		t.Errorf("round trip = %s; want %s", again, want)
	}
}

func TestMapJSONRejectsNonCanonicalKeys(t *testing.T) {
	for _, data := range []string{`{"0":"first","null":"last"}`, `{" 1":"x"}`, `{"1.0":"x"}`} {
		var m Map[uint32, string]
		if err := json.Unmarshal([]byte(data), &m); err == nil {
			t.Errorf("Unmarshal(%s) into Map[uint32, string] returned nil error", data)
		}
	}
	var b Map[bool, int]
	if err := json.Unmarshal([]byte(`{"false":1,"null":2}`), &b); err == nil {
		t.Error("Unmarshal with key null into Map[bool, int] returned nil error")
	}
	if err := json.Unmarshal([]byte(`{"false":1,"true":2}`), &b); err != nil {
		t.Fatal(err)
	}
	if v, _ := b.Get(false); v != 1 || b.Len() != 2 {
		t.Errorf("Get(false) = %d, Len() = %d; want 1, 2", v, b.Len())
	}
	var s Map[string, int]
	if err := json.Unmarshal([]byte(`{"null":1}`), &s); err != nil {
		t.Fatal(err)
	}
	if v, ok := s.Get("null"); !ok || v != 1 {
		t.Errorf(`Get("null") = %d, %t; want 1, true`, v, ok)
	}
}

func TestMapJSONNull(t *testing.T) {
	m := duplicateKeys()
	if err := json.Unmarshal([]byte(`null`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Len() != 2 {
		t.Errorf("Unmarshal(null) changed the map: Len() = %d", m.Len())
	}
	if err := json.Unmarshal([]byte(`[1]`), &m); err == nil {
		t.Error("Unmarshal([1]) returned nil error")
	}
}
