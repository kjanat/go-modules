//go:build !tinygo

package wit

import (
	"strings"
	"testing"
)

func TestUnsupportedParserFeatures(t *testing.T) {
	for _, tc := range []struct{ decl, kind string }{
		{"f: async func();", "async-freestanding"},
		{"type table = map<string, u32>;", "map"},
	} {
		src := "package test:features; interface api { " + tc.decl + " } world example { export api; }"
		r, err := DecodeWIT(strings.NewReader(src))
		if r != nil || err == nil || !strings.Contains(err.Error(), "unsupported WIT") || !strings.Contains(err.Error(), tc.kind) {
			t.Errorf("DecodeWIT(%s) = %v, %v; want unsupported %s", tc.decl, r, err, tc.kind)
		}
	}
}
