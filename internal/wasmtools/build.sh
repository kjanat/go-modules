#!/bin/sh
# Rebuild the checked-in WASI guest using upstream's workspace and Cargo.lock.
set -eu
cd "$(dirname "$0")"
rev=b38a9a674fb133984ce58e7ae853b07f3418e10e
sha256=9a4a9cdf5f62fb3a93cc6313b3e7382cd1b30d6a1de09a0cd3f4fc6bf4bdcc2b
mkdir -p target
tmp=$(mktemp -d target/source.XXXXXXXX)
trap 'rm -rf "$tmp"' EXIT
curl --fail --silent --show-error --location "https://codeload.github.com/bytecodealliance/wasm-tools/tar.gz/$rev" -o "$tmp/source.tar.gz"
printf '%s  %s\n' "$sha256" "$tmp/source.tar.gz" | sha256sum --check
tar -xzf "$tmp/source.tar.gz" -C "$tmp"
export CARGO_PROFILE_RELEASE_LTO=true CARGO_PROFILE_RELEASE_CODEGEN_UNITS=1
export CARGO_PROFILE_RELEASE_OPT_LEVEL=z CARGO_PROFILE_RELEASE_PANIC=abort CARGO_PROFILE_RELEASE_STRIP=true
export CARGO_TARGET_DIR="$PWD/target"
export CARGO_ENCODED_RUSTFLAGS
CARGO_ENCODED_RUSTFLAGS=$(printf '%s\037%s' "--remap-path-prefix=$PWD/$tmp/wasm-tools-$rev=/wasm-tools" "--remap-path-prefix=${CARGO_HOME:-$HOME/.cargo}/registry=/cargo/registry")
cargo +1.100.0 build --locked --manifest-path "$tmp/wasm-tools-$rev/Cargo.toml" \
  --target wasm32-wasip1 --release -p wasm-tools --no-default-features --features component
gzip -n -c "$CARGO_TARGET_DIR/wasm32-wasip1/release/wasm-tools.wasm" > "$tmp/guest.gz"
mv "$tmp/guest.gz" wasm-tools.wasm.gz
