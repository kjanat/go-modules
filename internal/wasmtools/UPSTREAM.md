# Embedded WIT parser

The committed wasm32-wasip1 executable is built from wasm-tools 1.263.0 at
[upstream revision b38a9a6](https://github.com/bytecodealliance/wasm-tools/commit/b38a9a674fb133984ce58e7ae853b07f3418e10e).
This revision includes the post-release parser depth and bounded-read fixes,
and checks section sizes against the remaining input before adding offsets.
The build uses the upstream Cargo.lock and only the component feature.

To rebuild on Linux, install Rust 1.100.0 with the wasm32-wasip1 target
(`rustup toolchain install 1.100.0 --profile minimal --target wasm32-wasip1`),
then run `make wasm-tools`. The build verifies the source archive before
compiling and replaces the gzip archive only on success.

To upgrade, update the revision and archive SHA-256 in build.sh, rebuild,
align the native wasm-tools version in CI, and run `make tests/generated`
with that native CLI on PATH. Run `go test ./... ./cm/...` against the
checked-in guest. The WIT suite covers parser limits, malformed sections,
and round trips from the fixture corpus.

The Go bindings implement a subset of the parser grammar. Unsupported function
and type kinds (for example async functions and maps) produce a decoder error.
