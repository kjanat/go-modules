package wit

import (
	"strings"
	"testing"
)

func TestMapDespecialize(t *testing.T) {
	value := &TypeDef{Kind: &Record{Fields: []Field{{Name: "id", Type: U32{}}}}}
	m := &Map{Key: String{}, Value: value}
	l, ok := Despecialize(m).(*List)
	if !ok {
		t.Fatalf("Despecialize returned %T; want *List", Despecialize(m))
	}
	td, ok := l.Type.(*TypeDef)
	if !ok {
		t.Fatalf("list element is %T; want *TypeDef", l.Type)
	}
	tuple, ok := td.Kind.(*Tuple)
	if !ok || len(tuple.Types) != 2 || tuple.Types[0] != m.Key || tuple.Types[1] != m.Value {
		t.Fatalf("list element kind is %#v; want tuple<string, value>", td.Kind)
	}
	if !DependsOn(m, value) {
		t.Error("map does not depend on its value type")
	}
	if !HasPointer(m) {
		t.Error("map has no pointer")
	}
}

func TestMalformedMapJSON(t *testing.T) {
	for _, pair := range []string{`[]`, `["string"]`, `["string",null]`, `["string","u32","bool"]`, `{}`} {
		t.Run(pair, func(t *testing.T) {
			input := `{"types":[{"kind":{"map":` + pair + `}}]}`
			_, err := DecodeJSON(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), "invalid WIT map") {
				t.Fatalf("DecodeJSON returned %v, want invalid map error", err)
			}
		})
	}
}
