package wit

// Map represents a WIT [map type] with key type K and value type V.
// It implements the [Node], [ABI], and [TypeDefKind] interfaces.
//
// [map type]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/WIT.md#types
type Map struct {
	_typeDefKind
	Key   Type
	Value Type
}

// Despecialize despecializes [Map] m into a [List] of [Tuple] with key and value types.
// See the [canonical ABI documentation] for more information.
//
// [canonical ABI documentation]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (m *Map) Despecialize() TypeDefKind {
	return &List{
		Type: &TypeDef{Kind: &Tuple{Types: []Type{m.Key, m.Value}}},
	}
}

// Size returns the [ABI byte size] for [Map] m.
// It is first [despecialized] into a [List] of [Tuple], then sized.
//
// [ABI byte size]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#element-size
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (m *Map) Size() uintptr {
	return m.Despecialize().Size()
}

// Align returns the [ABI byte alignment] for [Map] m.
// It is first [despecialized] into a [List] of [Tuple], then aligned.
//
// [ABI byte alignment]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#alignment
// [despecialized]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#despecialization
func (m *Map) Align() uintptr {
	return m.Despecialize().Align()
}

// Flat returns the [flattened] ABI representation of [Map] m.
//
// [flattened]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md#flattening
func (m *Map) Flat() []Type {
	return m.Despecialize().Flat()
}
