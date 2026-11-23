package wit

// Map represents a WIT map with distinct key and value types.
// Its memory layout is a pointer and length for a sequence of key/value pairs.
// The layout does not capture map-specific rules such as distinct keys;
// support for reading WIT maps does not imply support for generating bindings.
type Map struct {
	_typeDefKind
	Key   Type
	Value Type
}

// Size returns the Canonical ABI byte size of the pointer and length on wasm32.
func (*Map) Size() uintptr { return 8 }

// Align returns the Canonical ABI byte alignment of a map on wasm32.
func (*Map) Align() uintptr { return 4 }

// Flat returns the Canonical ABI pointer to the entries and the entry count.
func (m *Map) Flat() []Type {
	entry := &TypeDef{Kind: &Tuple{Types: []Type{m.Key, m.Value}}}
	return []Type{PointerTo(entry), U32{}}
}

func (*Map) hasPointer() bool    { return true }
func (m *Map) hasBorrow() bool   { return HasBorrow(m.Key) || HasBorrow(m.Value) }
func (m *Map) hasResource() bool { return HasResource(m.Key) || HasResource(m.Value) }
func (m *Map) dependsOn(dep Node) bool {
	return dep == m || DependsOn(m.Key, dep) || DependsOn(m.Value, dep)
}
