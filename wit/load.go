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
// cold compile completes first.
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
	wasmTools, err := wasmtools.New(ctx)
	if err != nil {
		return nil, errors.Join(err, ctx.Err())
	}
	defer wasmTools.Close(context.WithoutCancel(ctx))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err = wasmTools.Run(ctx, stdin, stdout, stderr, fsMap, args...)
	if ctx.Err() != nil {
		return nil, errors.Join(err, ctx.Err())
	}
	if err != nil {
		if _, ok := err.(*sys.ExitError); ok {
			return nil, fmt.Errorf("wasm-tools: %s", stderr.String())
		}
		return nil, fmt.Errorf("wasm-tools: %w", err)
	}
	return DecodeJSON(stdout)
}
