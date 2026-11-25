//go:build !wasip1 || !wasm || tinygo

package cm

// MapResult represents a WIT result that contains a map.
type MapResult[Shape, OK, Err any] = Result[Shape, OK, Err]
