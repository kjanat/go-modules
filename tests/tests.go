// Package tests exists as a standalone module in the wasm-tools-go repository for
// testing Go code generated from WIT using wit-bindgen-go.
// There are no user-importable packages in this module.

package tests

//go:generate rm -rf ./generated/*
//go:generate mkdir -p ./generated
//go:generate go run github.com/kjanat/go-modules/cmd/wit-bindgen-go generate --versioned -o ./generated ../testdata/wasi/cli.wit.json
//go:generate mkdir -p ./maps ./numeric
//go:generate find ./maps ./numeric -type f ! -name *_test.go -delete
//go:generate go run github.com/kjanat/go-modules/cmd/wit-bindgen-go generate -o ./maps ../testdata/codegen/maps.wit.json
//go:generate go run github.com/kjanat/go-modules/cmd/wit-bindgen-go generate -o ./numeric ../testdata/codegen/map-numeric-input.wit.json
