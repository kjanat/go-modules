//go:build !tinygo

package wit

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kjanat/wazero/experimental"
)

func TestLoadWITPaths(t *testing.T) {
	// Relative paths require the fixture and working directory to share a volume.
	root, err := os.MkdirTemp(".", ".wit-load-paths-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	root, err = filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "nested", "example.wit")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("package example:paths; world example {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, file)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct {
		name string
		path string
	}{
		{"absolute", file},
		{"absolute slash", filepath.ToSlash(file)},
		{"relative", rel},
		{"relative slash", filepath.ToSlash(rel)},
	} {
		t.Run(input.name, func(t *testing.T) {
			resolved, err := LoadWIT(input.path)
			if err != nil {
				t.Fatal(err)
			}
			if len(resolved.Worlds) != 1 || resolved.Worlds[0].Name != "example" {
				t.Fatalf("resolved worlds: %v", resolved.Worlds)
			}
		})
	}
}

func TestLoadWITContextCancelled(t *testing.T) {
	for _, expired := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		want := context.Canceled
		if expired {
			cancel()
			ctx, cancel = context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
			want = context.DeadlineExceeded
		}
		cancel()
		for _, load := range []func() (*Resolve, error){
			func() (*Resolve, error) { return LoadWITContext(ctx, "missing.wit") },
			func() (*Resolve, error) {
				return DecodeWITContext(ctx, strings.NewReader("package example:cancel; world example {}"))
			},
		} {
			resolved, err := load()
			if resolved != nil || !errors.Is(err, want) {
				t.Fatalf("got (%v, %v), want cancellation %v", resolved, err, want)
			}
		}
	}
}

func TestDecodeWITClosesRuntimeAfterCancelledInput(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	// The guest can close itself during Run. Its WASI provider stays alive
	// until the loader closes the runtime with the detached request context.
	type key struct{}
	ctx = context.WithValue(ctx, key{}, "request")
	var detachedCloses int
	ctx = experimental.WithCloseNotifier(ctx, experimental.CloseNotifyFunc(func(closeCtx context.Context, code uint32) {
		if code == 0 && closeCtx.Done() == nil && closeCtx.Value(key{}) == "request" {
			detachedCloses++
		}
	}))
	reader := &cancellingWITReader{Reader: strings.NewReader("package example:cancel; world example {}"), cancel: cancel}
	resolved, err := DecodeWITContext(ctx, reader)
	if resolved != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("got (%v, %v), want cancellation after the guest reads stdin", resolved, err)
	}
	// If the small guest finishes before wazero observes cancellation, both
	// it and the provider can still be open at runtime teardown.
	if detachedCloses == 0 {
		t.Fatal("runtime teardown did not close any module with the detached request context")
	}
	// Cancellation must leave subsequent, independent loads usable.
	resolved, err = DecodeWITContext(t.Context(), strings.NewReader("package example:cancel; world example {}"))
	if err != nil || len(resolved.Worlds) != 1 || resolved.Worlds[0].Name != "example" {
		t.Fatalf("load after cancellation: %v, %v", resolved, err)
	}
}

type cancellingWITReader struct {
	io.Reader
	cancel context.CancelFunc
}

func (r *cancellingWITReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.cancel()
	return n, err
}

func TestLoadWITDirectoryDependencies(t *testing.T) {
	root := t.TempDir()
	dependency := filepath.Join(root, "deps", "dep")
	if err := os.MkdirAll(dependency, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{
		filepath.Join(root, "main.wit"):      "package example:paths; world example { import example:dep/api; }",
		filepath.Join(dependency, "api.wit"): "package example:dep; interface api { ping: func(); }",
	} {
		if err := os.WriteFile(name, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	resolved, err := LoadWIT(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Packages) != 2 || len(resolved.Worlds) != 1 || resolved.Worlds[0].Name != "example" {
		t.Fatalf("resolved packages/worlds: %v / %v", resolved.Packages, resolved.Worlds)
	}
}

func TestDecodeWITOutputCancellation(t *testing.T) {
	for _, tc := range []struct{ name, input string }{
		{"last buffered read", "{}"}, {"before refill", strings.Repeat(" ", 8192) + "{}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			source := strings.NewReader(tc.input)
			res, err := decodeWITOutput(ctx, &cancellingWITReader{Reader: source, cancel: cancel})
			if res != nil || !errors.Is(err, context.Canceled) {
				t.Fatalf("got (%v, %v), want cancellation", res, err)
			}
			if tc.name == "before refill" && source.Len() == 0 {
				t.Fatal("decoder consumed the remaining input after cancellation")
			}
		})
	}
}

func TestDecodeWITCancelledWhileClosingAfterGuestError(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	inClose := false
	ctx = experimental.WithCloseNotifier(ctx, experimental.CloseNotifyFunc(func(closeCtx context.Context, _ uint32) {
		// The loader detaches the context for runtime teardown, after Run
		// has returned the guest's stderr. Ignore earlier module closes.
		if closeCtx.Done() == nil {
			inClose = true
			cancel()
		}
	}))
	res, err := DecodeWITContext(ctx, strings.NewReader("package example:broken; world ?"))
	if !inClose || res != nil || !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "world ?") {
		t.Fatalf("cancelled in close=%v, result=(%v, %v); want guest diagnostic and cancellation", inClose, res, err)
	}
}
