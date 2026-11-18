package wit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjanat/go-modules/internal/ctxerr"
	"github.com/kjanat/go-modules/internal/wasmtools"
	"github.com/kjanat/wazero/sys"
)

// LoadJSON loads a [WIT] JSON file from path.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
func LoadJSON(path string) (*Resolve, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DecodeJSON(f)
}

// LoadWIT loads [WIT] data from path using the embedded [wasm-tools].
// Use [LoadWITContext] to cancel guest execution.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
// [wasm-tools]: https://crates.io/crates/wasm-tools
func LoadWIT(path string) (*Resolve, error) {
	return LoadWITContext(context.Background(), path)
}

// LoadWITContext loads WIT data from path using the embedded wasm-tools.
// Cancelling ctx interrupts guest execution. Compilation of the embedded
// wasm-tools module does not observe ctx, so a cold compile completes first.
// The runtime is closed before decoding its JSON output. The decoder checks
// cancellation between input reads and before returning.
func LoadWITContext(ctx context.Context, path string) (*Resolve, error) {
	return loadWIT(ctx, path, nil)
}

// DecodeWIT decodes [WIT] data from Reader r using the embedded [wasm-tools].
// Use [DecodeWITContext] to cancel guest execution.
//
// [WIT]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md
// [wasm-tools]: https://crates.io/crates/wasm-tools
func DecodeWIT(r io.Reader) (*Resolve, error) {
	return DecodeWITContext(context.Background(), r)
}

// DecodeWITContext decodes WIT data from r using the embedded wasm-tools.
// Cancelling ctx interrupts guest execution. A blocking Read
// must be interrupted separately, for example by closing a caller-owned reader.
// Compilation of the embedded wasm-tools module does not observe ctx, so a
// cold compile completes first. JSON output has the same cancellation
// boundaries as [LoadWITContext].
func DecodeWITContext(ctx context.Context, r io.Reader) (*Resolve, error) {
	return loadWIT(ctx, "", r)
}

// loadWIT loads WIT data from path or reader by processing it through wasm-tools.
// It accepts either a path or an io.Reader as input, but not both.
// If the path is not "" and "-", it will be used as the input file.
// Otherwise, the reader will be used as the input.
func loadWIT(ctx context.Context, path string, reader io.Reader) (*Resolve, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if path != "" && reader != nil {
		return nil, errors.New("cannot set both path and reader; provide only one")
	}

	args := []string{"component", "wit", "-j", "--all-features"}
	fsMap := make(map[string]fs.FS)
	var stdin io.Reader

	if path != "" {
		path = strings.TrimPrefix(path, "./")
		args = append(args, filepath.ToSlash(path))
		dir := filepath.Dir(path)
		fsMap[filepath.ToSlash(dir)] = os.DirFS(dir)
	} else {
		stdin = reader
	}
	var stdout bytes.Buffer
	// Runtime teardown can cancel the request even after the guest failed.
	err := runWIT(ctx, stdin, &stdout, fsMap, args)
	if err := ctxerr.Join(ctx, err); err != nil {
		return nil, err
	}
	return decodeWITOutput(ctx, &stdout)
}

// runWIT owns the runtime only while producing JSON, including on panic.
func runWIT(ctx context.Context, stdin io.Reader, stdout io.Writer, fsMap map[string]fs.FS, args []string) (err error) {
	tool, err := wasmtools.New(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := tool.Close(context.WithoutCancel(ctx)); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing wasm-tools runtime: %w", closeErr))
		}
	}()
	var stderr bytes.Buffer
	err = tool.Run(ctx, stdin, stdout, &stderr, fsMap, args...)
	if err != nil {
		if exit, ok := err.(*sys.ExitError); ok && !errors.Is(exit, context.Canceled) && !errors.Is(exit, context.DeadlineExceeded) {
			return fmt.Errorf("wasm-tools: %s", stderr.String())
		}
		return fmt.Errorf("wasm-tools: %w", err)
	}
	return nil
}

func decodeWITOutput(ctx context.Context, r io.Reader) (*Resolve, error) {
	res, err := DecodeJSON(contextReader{ctx, r})
	// The decoder may finish entirely from its last buffered read.
	if err := ctxerr.Join(ctx, err); err != nil {
		return nil, err
	}
	return res, nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
