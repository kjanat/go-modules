# Embedded WIT parser

The executable uses wasm-tools 1.232.0 with a local wit-parser 0.232.0 patch.
The parser source comes from the published `wit-parser-0.232.0.crate`; the
vendored copy retains the library source, manifest, README and upstream licenses.
The Cargo manifest uses a path override so the executable and its source match.

The backport follows [wasm-tools #2813](https://github.com/bytecodealliance/wasm-tools/commit/b38a9a674fb133984ce58e7ae853b07f3418e10e):

- `src/decoding.rs` grows the section buffer with bytes actually read instead
  of allocating the length claimed by the input.
- `src/resolve.rs` checks named-type and alias depth in topological order,
  without recursively walking the graph. Newly resolved types 100 levels deep
  are rejected.
- `src/ast.rs` also bounds inline parsing at 100 levels because 0.232.0 predates
  the inline-depth check assumed by the upstream change.

The vendored wasmparser 0.232.0 also uses subtraction in
`BinaryReader::ensure_has_bytes`. On wasm32, adding a claimed section length to
the cursor could wrap and turn truncated input into a slice panic before the
WIT decoder could return its error. The malformed-section fixture covers that
path as well as bounded buffering.

The older parser preserves the generator's supported WIT syntax. An upgrade to
the November 2026 wasm-tools source rejects existing empty-record, empty-tuple
and resource-name fixtures. Moving to that syntax requires a separate migration.

Rebuild from the repository root with Rust's `wasm32-wasip1` target installed:

```sh
make wasm-tools
go test ./... ./cm/...
```

The Cargo lockfile fixes dependency versions. `gzip -n` omits timestamps and
input filenames from the committed executable archive. `wit/parser_limits_test.go`
exercises the embedded executable through the public loader.
