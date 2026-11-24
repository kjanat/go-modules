#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../tests"
scratch=$(mktemp -d)
trap 'rm -rf -- "$scratch"' EXIT

adapter=${WASI_REACTOR_ADAPTER:-}
go_bin=${WIT_MAP_TEST_GO:-go}
if [[ -z "$adapter" ]]; then
	read -r _ wasmtime_version _ <<<"$(wasmtime --version)"
	adapter="$scratch/wasi_snapshot_preview1.reactor.wasm"
	gh release download "v$wasmtime_version" --repo bytecodealliance/wasmtime \
		--pattern wasi_snapshot_preview1.reactor.wasm --dir "$scratch"
fi

sed '/^  import maps;/d' ../testdata/codegen/maps.wit >"$scratch/maps.wit"
cat >>"$scratch/maps.wit" <<'WIT'
world consumer {
  import maps;
  export check: func() -> u32;
}
WIT

build_component() {
	local world=$1 guest=$2
	GOWORK=off GOOS=wasip1 GOARCH=wasm "$go_bin" build -ldflags=-checklinkname=0 \
		-buildmode=c-shared -o "$scratch/$guest.core.wasm" "./guest/maps-go/$guest"
	wasm-tools component embed --world "$world" "$scratch/maps.wit" \
		"$scratch/$guest.core.wasm" -o "$scratch/$guest.embedded.wasm"
	wasm-tools component new "$scratch/$guest.embedded.wasm" \
		--adapt "wasi_snapshot_preview1=$adapter" -o "$scratch/$guest.wasm"
}

build_component the-maps provider
build_component consumer consumer
wasm-tools compose "$scratch/consumer.wasm" -d "$scratch/provider.wasm" -o "$scratch/linked.wasm"
wasm-tools validate --features all "$scratch/linked.wasm"

check() {
	local component=$1 invocation=$2 expected=$3 actual
	actual=$(wasmtime run -W component-model-map=y --invoke "$invocation" "$scratch/$component.wasm")
	printf '%s => %s\n' "$invocation" "$actual"
	[[ "$actual" == "$expected" ]]
}
check provider 'named-roundtrip({"x": 1, "x": 2})' '{"x": 2}'
check provider 'same-roundtrip({})' '{}'
check linked 'check()' 42
