package cm

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"strings"
	"testing"
)

type point struct{ N int }

func TestMapJSONv2CallerMarshalers(t *testing.T) {
	m := ToMap(map[string]point{"x": {N: 42}})
	opts := json.WithMarshalers(json.MarshalFunc(func(point) ([]byte, error) {
		return []byte(`"custom"`), nil
	}))
	got, err := json.Marshal(m, opts)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"x":"custom"}`; string(got) != want {
		t.Errorf("Marshal = %s; want %s", got, want)
	}
}

func TestMapJSONv2CallerUnmarshalers(t *testing.T) {
	opts := json.WithUnmarshalers(json.UnmarshalFunc(func(b []byte, p *point) error {
		p.N = len(b)
		return nil
	}))
	var m Map[string, point]
	if err := json.Unmarshal([]byte(`{"x":"abc"}`), &m, opts); err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Get("x"); v.N != 5 {
		t.Errorf("Get(x).N = %d; want 5", v.N)
	}
}

func TestMapJSONv2DuplicateNames(t *testing.T) {
	data := []byte(`{"x":1,"x":2}`)
	var m Map[string, uint32]
	if err := json.Unmarshal(data, &m); err == nil {
		t.Error("Unmarshal with duplicate names returned nil error under default v2 options")
	}
	if err := json.Unmarshal(data, &m, jsontext.AllowDuplicateNames(true)); err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Get("x"); v != 2 || m.Len() != 1 {
		t.Errorf("Get(x) = %d, Len() = %d; want 2, 1", v, m.Len())
	}
	if _, n := LowerMap[Tuple[string, uint32]](m); n != 2 {
		t.Errorf("Unmarshal kept %d pairs; want 2", n)
	}
}

func TestMapJSONv2RejectsNullKey(t *testing.T) {
	var m Map[uint32, uint32]
	if err := json.Unmarshal([]byte(`{"0":1,"null":2}`), &m); err == nil {
		t.Error("Unmarshal with key null into Map[uint32, uint32] returned nil error")
	}
}

type label string

func TestMapJSONv2CallerKeyCodecs(t *testing.T) {
	marshal := json.WithMarshalers(json.MarshalFunc(func(l label) ([]byte, error) {
		return json.Marshal("encoded-" + string(l))
	}))
	got, err := json.Marshal(ToMap(map[label]int{"original": 1}), marshal)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"encoded-original":1}`; string(got) != want {
		t.Errorf("Marshal = %s; want %s", got, want)
	}

	unmarshal := json.WithUnmarshalers(json.UnmarshalFunc(func(b []byte, l *label) error {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*l = label(strings.TrimPrefix(s, "encoded-"))
		return nil
	}))
	var m Map[label, int]
	if err := json.Unmarshal(got, &m, unmarshal); err != nil {
		t.Fatal(err)
	}
	if v, ok := m.Get("original"); !ok || v != 1 {
		t.Errorf(`Get("original") = %d, %t; want 1, true`, v, ok)
	}
}
