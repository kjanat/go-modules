//go:build !tinygo

package wit

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestDecodeWITTypeDepth(t *testing.T) {
	for _, shape := range []string{"aliases", "records", "inline"} {
		for _, depth := range []int{80, 120} {
			t.Run(fmt.Sprintf("%s/%d", shape, depth), func(t *testing.T) {
				var text strings.Builder
				text.WriteString("package example:limits; interface types {\n")
				if shape == "inline" {
					fmt.Fprintf(&text, "type nested = %su32%s;\n", strings.Repeat("list<", depth), strings.Repeat(">", depth))
				} else {
					text.WriteString("type t0 = u32;\n")
					for i := 1; i < depth; i++ {
						if shape == "records" {
							fmt.Fprintf(&text, "record t%d { x: t%d }\n", i, i-1)
						} else {
							fmt.Fprintf(&text, "type t%d = t%d;\n", i, i-1)
						}
					}
				}
				text.WriteString("} world example { export types; }")
				resolved, err := DecodeWIT(strings.NewReader(text.String()))
				if depth < 100 {
					if err != nil || resolved == nil {
						t.Fatalf("valid depth rejected: %v", err)
					}
				} else if resolved != nil || err == nil || !strings.Contains(err.Error(), "too deep") {
					t.Fatalf("expected a nesting error, got %v, %v", resolved, err)
				}
			})
		}
	}
}

func TestDecodeWITTruncatedLargeSection(t *testing.T) {
	// A component header followed by a custom section claiming 4 GiB - 1 bytes.
	input := []byte{0, 'a', 's', 'm', 0x0d, 0, 1, 0, 0, 0xff, 0xff, 0xff, 0xff, 0x0f}
	resolved, err := DecodeWIT(bytes.NewReader(input))
	if resolved != nil || err == nil || !strings.Contains(err.Error(), "unexpected end-of-file") {
		t.Fatalf("expected truncated input without a guest allocation trap, got %v, %v", resolved, err)
	}
}
