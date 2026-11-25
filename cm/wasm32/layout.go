//go:build wasip1 && wasm && !tinygo

package wasm32

// Layout describes the relationship between a value in linear memory and its
// flattened core Wasm arguments. A variant has a tag field followed by the
// flattened fields of the selected case, with the remaining slots zeroed.
type Layout struct {
	Flat   int
	Scalar uint8
	Fields []Field
	Cases  []Field
}

type Field struct {
	Offset uint32
	Layout *Layout
}

// Signed marks a scalar that must be sign-extended into an i32.
const Signed = 16

func (l *Layout) Flatten(addr uint32, flat []uint64) {
	if l.Scalar != 0 {
		flat[0] = loadScalar(l.Scalar, addr)
		return
	}
	if l.Cases != nil {
		l.Fields[0].Layout.Flatten(addr, flat)
		clear(flat[1:])
		c := l.Cases[flat[0]]
		if c.Layout != nil {
			c.Layout.Flatten(addr+c.Offset, flat[1:])
		}
		return
	}
	for _, f := range l.Fields {
		f.Layout.Flatten(addr+f.Offset, flat)
		flat = flat[f.Layout.Flat:]
	}
}

func (l *Layout) Unflatten(addr uint32, flat []uint64) {
	if l.Scalar != 0 {
		storeScalar(l.Scalar, addr, flat[0])
		return
	}
	if l.Cases != nil {
		l.Fields[0].Layout.Unflatten(addr, flat)
		c := l.Cases[flat[0]]
		if c.Layout != nil {
			c.Layout.Unflatten(addr+c.Offset, flat[1:])
		}
		return
	}
	for _, f := range l.Fields {
		f.Layout.Unflatten(addr+f.Offset, flat)
		flat = flat[f.Layout.Flat:]
	}
}

func loadScalar(kind uint8, addr uint32) uint64 {
	switch kind {
	case 1:
		return uint64(Load[uint8](addr))
	case 2:
		return uint64(Load[uint16](addr))
	case 4:
		return uint64(Load[uint32](addr))
	case 8:
		return Load[uint64](addr)
	case Signed | 1:
		return uint64(uint32(int32(Load[int8](addr))))
	case Signed | 2:
		return uint64(uint32(int32(Load[int16](addr))))
	}
	panic("invalid canonical ABI scalar")
}

func storeScalar(kind uint8, addr uint32, value uint64) {
	switch kind &^ Signed {
	case 1:
		Store(addr, uint8(value))
	case 2:
		Store(addr, uint16(value))
	case 4:
		Store(addr, uint32(value))
	case 8:
		Store(addr, value)
	default:
		panic("invalid canonical ABI scalar")
	}
}
