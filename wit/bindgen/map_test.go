//go:build !tinygo

package bindgen

import (
	"strings"
	"testing"

	"github.com/kjanat/go-modules/wit"
)

func TestGoChecksMapsInSelectedWorld(t *testing.T) {
	const src = `package test:features;
interface api {
	echo: func(input: option<map<string, u32>>);
}
interface named {
	type table = map<string, u32>;
	get: func() -> table;
}
world example { export api; }
world tables { export named; }
world simple { export echo: func(input: string) -> string; }
`
	res, err := wit.DecodeWIT(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		opts []Option
		want string
	}{
		{"first world", nil, `"(anonymous)" in world "example"`},
		{"anonymous map", []Option{World("example")}, `"(anonymous)" in world "example"`},
		{"named map", []Option{World("tables")}, `"table" in world "tables"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pkgs, err := Go(res, tc.opts...)
			if pkgs != nil || err == nil || !strings.Contains(err.Error(), "does not support WIT map types: "+tc.want) {
				t.Fatalf("Go returned %v, %v; want no packages and a WIT map error for %s", pkgs, err, tc.want)
			}
		})
	}
	pkgs, err := Go(res, World("simple"))
	if len(pkgs) == 0 || err != nil {
		t.Fatalf("Go returned %v, %v; want bindings for the map-free world", pkgs, err)
	}
}

func TestGoWithoutWorldSkipsMapCheck(t *testing.T) {
	const src = `package test:features;
interface api {
	type table = map<string, u32>;
}
`
	res, err := wit.DecodeWIT(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Go(res)
	if err != nil {
		t.Fatalf("Go returned %v, %v; want no error without a world", pkgs, err)
	}
}
