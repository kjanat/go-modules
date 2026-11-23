package wit

import (
	"strings"
	"testing"
)

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
