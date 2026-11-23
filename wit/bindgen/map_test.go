//go:build !tinygo

package bindgen

import (
	"strings"
	"testing"

	"github.com/kjanat/go-modules/wit"
)

func TestGoRejectsNestedAnonymousMap(t *testing.T) {
	const src = `package test:features;
interface api {
	echo: func(input: option<map<string, u32>>);
}
world example { export api; }
`
	res, err := wit.DecodeWIT(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Go(res)
	if pkgs != nil || err == nil || !strings.Contains(err.Error(), "does not support WIT map types") {
		t.Fatalf("Go returned %v, %v; want no packages and a WIT map error", pkgs, err)
	}
}
