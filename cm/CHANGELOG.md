# Changelog

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.4.0] - 2026-11-24

### Added

- Module `github.com/kjanat/go-modules/cm`. Its types are aliases of the `go.bytecodealliance.org/cm` v0.3.0 types, and its functions call the upstream functions.
- `Map[K, V]` represents WIT `map<K, V>` with the memory layout of `List[Tuple[K, V]]`. `Get`, `All`, `Len`, and `Map` resolve duplicate keys to the last value. `ToMap` builds a `Map` from a Go map.
- `LowerMap` and `LiftMap` convert a `Map` to and from its ABI pointer and pair count.
- `Map` marshals to and from a JSON object through `encoding/json/v2`. Caller options apply to its values, and the caller's decoder decides whether duplicate names are allowed.

[v0.4.0]: https://github.com/kjanat/go-modules/tree/cm/v0.4.0
