//go:build !tinygo

package witcli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const exampleWIT = "package example:cancel; world example {}"

func TestLoadWITForwardsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	file := filepath.Join(t.TempDir(), "example.wit")
	if err := os.WriteFile(file, []byte(exampleWIT), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"-", file} {
		res, err := LoadWIT(ctx, path, strings.NewReader(exampleWIT), path == "-")
		if res != nil || !errors.Is(err, context.Canceled) {
			t.Errorf("%q: got (%v, %v), want cancellation", path, res, err)
		}
	}
}

func TestLoadOCICancelsBetweenPullAndDecode(t *testing.T) {
	const ref = "registry.example/test/wit:v1"
	if runtime.GOOS == "wasip1" || runtime.GOOS == "wasip2" {
		t.Skip("OCI not supported on this target")
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	called := false
	res, err := loadWIT(ctx, ref, nil, false, func(fetchCtx context.Context, path string) ([]byte, error) {
		called = true
		if fetchCtx != ctx || path != ref {
			t.Fatal("pull did not receive the caller context and reference")
		}
		cancel()
		return []byte(exampleWIT), nil
	})
	if !called || res != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("pull=%v, result=(%v, %v); want cancellation after pull", called, res, err)
	}
}

// Exercise the real registry client's error wrapping as well as the successful
// handoff above. A pre-cancelled pull requires no live registry.
func TestLoadOCIPreservesPullCancellation(t *testing.T) {
	if runtime.GOOS == "wasip1" || runtime.GOOS == "wasip2" {
		t.Skip("OCI not supported on this target")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	res, err := LoadWIT(ctx, "127.0.0.1:1/test/wit:v1", nil, false)
	if res != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled OCI pull: got (%v, %v)", res, err)
	}
}
