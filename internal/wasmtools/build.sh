#!/bin/sh
set -eu
cd "$(dirname "$0")"
export CARGO_TARGET_DIR="$PWD/target"
export CARGO_ENCODED_RUSTFLAGS
CARGO_ENCODED_RUSTFLAGS=$(printf '%s\037%s' "--remap-path-prefix=${CARGO_HOME:-$HOME/.cargo}=/cargo" "--remap-path-prefix=$PWD=/build")
cargo build --locked --release --target wasm32-wasip1 -p wasm-tools
tmp=$(mktemp wasm-tools.wasm.gz.XXXXXX)
trap 'rm -f "$tmp"' EXIT
gzip -n -c "$CARGO_TARGET_DIR/wasm32-wasip1/release/wasm-tools.wasm" > "$tmp"
mv "$tmp" wasm-tools.wasm.gz
