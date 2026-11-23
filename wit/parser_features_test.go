//go:build !tinygo

package wit

import (
	"strings"
	"testing"
)

func TestUnsupportedParserFeatures(t *testing.T) {
	for _, tc := range []struct{ decl, kind string }{
		{"f: async func();", "async-freestanding"},
	} {
		src := "package test:features; interface api { " + tc.decl + " } world example { export api; }"
		r, err := DecodeWIT(strings.NewReader(src))
		if r != nil || err == nil || !strings.Contains(err.Error(), "unsupported WIT") || !strings.Contains(err.Error(), tc.kind) {
			t.Errorf("DecodeWIT(%s) = %v, %v; want unsupported %s", tc.decl, r, err, tc.kind)
		}
	}
}

func TestDecodeWITMapRoundTrip(t *testing.T) {
	src := `package test:features;
interface types {
	type value = u32;
	type %map = map<string, value>;
}
interface api {
	use types.{%map as table};
	echo: func(input: table) -> table;
}
world example { export api; }
`

	check := func(text string) *Resolve {
		t.Helper()
		res, err := DecodeWIT(strings.NewReader(text))
		if err != nil {
			t.Fatal(err)
		}
		var found bool
		for _, td := range res.TypeDefs {
			if td.TypeName() != "table" {
				continue
			}
			root := td.Root()
			m, ok := root.Kind.(*Map)
			if !ok {
				t.Fatalf("table has kind %T; want map", root.Kind)
			}
			if root.TypeName() != "map" || m.Key != (String{}) || m.Value.TypeName() != "value" {
				t.Fatalf("table aliases %q with key %v and value %v", root.TypeName(), m.Key, m.Value)
			}
			if td.Owner == root.Owner {
				t.Fatal("table must import a map from another interface")
			}
			found = true
		}
		if !found {
			t.Fatal("could not find imported table")
		}
		return res
	}

	printed := check(src).WIT(nil, "")
	if !strings.Contains(printed, "type %map = map<string, value>;") ||
		!strings.Contains(printed, "use types.{%map as table};") {
		t.Fatalf("map or import missing from printed WIT:\n%s", printed)
	}
	check(printed)
}
