package bindgen

import (
	"fmt"
	"strings"

	"github.com/kjanat/go-modules/wit"
)

type wireField struct {
	name   string
	typ    wit.Type
	offset uintptr
}

func wireFields(kind wit.TypeDefKind) []wireField {
	switch k := kind.(type) {
	case *wit.Record:
		var fields []wireField
		var offset uintptr
		for _, field := range k.Fields {
			offset = wit.Align(offset, field.Type.Align())
			fields = append(fields, wireField{fieldName(field.Name, true), field.Type, offset})
			offset += field.Type.Size()
		}
		return fields
	case *wit.Tuple:
		var fields []wireField
		var offset uintptr
		for i, t := range k.Types {
			offset = wit.Align(offset, t.Align())
			name := fmt.Sprintf(".F%d", i)
			if k.Type() != nil {
				name = fmt.Sprintf("[%d]", i)
			}
			fields = append(fields, wireField{name, t, offset})
			offset += t.Size()
		}
		return fields
	}
	return nil
}

func wireKind(t wit.Type) wit.TypeDefKind {
	if td, ok := t.(*wit.TypeDef); ok {
		return td.Root().Kind
	}
	return t
}

func (w *wasm32Generator) codec(dir wit.Direction, t wit.Type) string {
	// An interface imported and exported in the same world shares the type
	// representation selected by typeRep for the requested direction.
	cacheKey := wireCodecKey{dir: dir, typ: t}
	if name, ok := w.codecs[cacheKey]; ok {
		return name
	}
	name := w.file.DeclareName("memory32Value")
	w.codecs[cacheKey] = name
	rep := w.g.typeRep(w.file, dir, t)
	cmName := func() string { return w.file.Import(w.g.opts.cmPackage) }
	var load, store strings.Builder
	switch k := wireKind(t).(type) {
	case wit.String:
		fmt.Fprintf(&load, "return %s(%s.LoadString(addr))\n", rep, w.wire)
		fmt.Fprintf(&store, "%s.StoreString(a, addr, string(v))\n", w.wire)
	case *wit.Record, *wit.Tuple:
		for _, f := range wireFields(k) {
			child := w.codec(dir, f.typ)
			field := f.name
			if !strings.HasPrefix(field, ".") && !strings.HasPrefix(field, "[") {
				field = "." + field
			}
			fmt.Fprintf(&load, "v%s = %sLoad(addr+%d)\n", field, child, f.offset)
			fmt.Fprintf(&store, "%sStore(a, addr+%d, v%s)\n", child, f.offset, field)
		}
		load.WriteString("return\n")
	case *wit.List:
		child := w.codec(dir, k.Type)
		fmt.Fprintf(&load, "return %s(%s.ToList(%s.LoadSequence(addr, %d, %sLoad)))\n", rep, cmName(), w.wire, k.Type.Size(), child)
		fmt.Fprintf(&store, "%s.StoreSequence(a, addr, %d, %d, v.Slice(), %sStore)\n", w.wire, k.Type.Size(), k.Type.Align(), child)
	case *wit.Map:
		key, value := w.codec(dir, k.Key), w.codec(dir, k.Value)
		entry := k.Despecialize().(*wit.List).Type
		pair := fmt.Sprintf("%s.Tuple[%s, %s]", cmName(), w.g.typeRep(w.file, dir, k.Key), w.g.typeRep(w.file, dir, k.Value))
		offset := wit.Align(k.Key.Size(), k.Value.Align())
		fmt.Fprintf(&load, "return %s.LoadMap[%s](addr, %d, func(p uint32) %s {\nreturn %s{F0: %sLoad(p), F1: %sLoad(p+%d)}\n})\n", w.wire, rep, entry.Size(), pair, pair, key, value, offset)
		fmt.Fprintf(&store, "%s.StoreMap(a, addr, %d, %d, v, func(a *%s.Arena, p uint32, pair %s) {\n%sStore(a, p, pair.F0)\n%sStore(a, p+%d, pair.F1)\n})\n", w.wire, entry.Size(), entry.Align(), w.wire, pair, key, value, offset)
	case *wit.Option:
		payload := w.codec(dir, k.Type)
		offset := wit.Align(1, k.Type.Align())
		fmt.Fprintf(&load, "switch %s.Load[uint8](addr) {\ncase 0: return\ncase 1: return %s(%s.Some(%sLoad(addr+%d)))\n}\npanic(\"invalid canonical option tag\")\n", w.wire, rep, cmName(), payload, offset)
		fmt.Fprintf(&store, "if some := v.Some(); some != nil {\n%s.Store(addr, uint8(1))\n%sStore(a, addr+%d, *some)\n} else { %s.Store(addr, uint8(0)) }\n", w.wire, payload, offset, w.wire)
	case *wit.Result:
		if k.OK == nil && k.Err == nil {
			w.scalarCodec(&load, &store, rep)
			break
		}
		cases := k.Despecialize().(*wit.Variant)
		fmt.Fprintf(&load, "switch %s.Load[uint8](addr) {\n", w.wire)
		store.WriteString("if v.IsOK() {\n")
		for i, t := range []wit.Type{k.OK, k.Err} {
			method := []string{"OK", "Err"}[i]
			expr := "struct{}{}"
			fmt.Fprintf(&store, "%s.Store(addr, uint8(%d))\n", w.wire, i)
			if t != nil {
				p := w.codec(dir, t)
				offset := payloadOffset(cases)
				expr = fmt.Sprintf("%sLoad(addr+%d)", p, offset)
				fmt.Fprintf(&store, "%sStore(a, addr+%d, *v.%s())\n", p, offset, method)
			}
			fmt.Fprintf(&load, "case %d: return %s.%s[%s](%s)\n", i, cmName(), method, rep, expr)
			if i == 0 {
				store.WriteString("} else {\n")
			}
		}
		load.WriteString("}\npanic(\"invalid canonical result tag\")\n")
		store.WriteString("}\n")
	case *wit.Variant:
		if k.Enum() != nil {
			w.scalarCodec(&load, &store, rep)
			break
		}
		tag := w.g.typeRep(w.file, dir, wit.Discriminant(len(k.Cases)))
		fmt.Fprintf(&load, "switch %s.Load[%s](addr) {\n", w.wire, tag)
		fmt.Fprintf(&store, "tag := v.Tag()\n%s.Store(addr, tag)\nswitch tag {\n", w.wire)
		for i, c := range k.Cases {
			expr := "struct{}{}"
			if c.Type != nil {
				payload := w.codec(dir, c.Type)
				offset := payloadOffset(k)
				expr = fmt.Sprintf("%sLoad(addr+%d)", payload, offset)
				fmt.Fprintf(&store, "case %d: %sStore(a, addr+%d, *%s.Case[%s](&v, %d))\n", i, payload, offset, cmName(), w.g.typeRep(w.file, dir, c.Type), i)
			}
			fmt.Fprintf(&load, "case %d: return %s.New[%s](%s(%d), %s)\n", i, cmName(), rep, tag, i, expr)
		}
		load.WriteString("}\npanic(\"invalid canonical variant tag\")\n")
		store.WriteString("}\n")
	default:
		w.scalarCodec(&load, &store, rep)
	}
	fmt.Fprintf(&w.declarations, "func %sLoad(addr uint32) (v %s) {\n%s}\nfunc %sStore(a *%s.Arena, addr uint32, v %s) {\n%s}\n\n", name, rep, load.String(), name, w.wire, rep, store.String())
	return name
}

func (w *wasm32Generator) scalarCodec(load, store *strings.Builder, rep string) {
	fmt.Fprintf(load, "return %s.Load[%s](addr)\n", w.wire, rep)
	fmt.Fprintf(store, "%s.Store(addr, v)\n", w.wire)
}

func payloadOffset(v *wit.Variant) uintptr {
	align := uintptr(1)
	for _, c := range v.Cases {
		if c.Type != nil {
			align = max(align, c.Type.Align())
		}
	}
	return wit.Align(wit.Discriminant(len(v.Cases)).Size(), align)
}

func (w *wasm32Generator) layout(t wit.Type) string {
	if name, ok := w.layouts[t]; ok {
		return name
	}
	name := w.file.DeclareName("memory32Layout")
	w.layouts[t] = name
	var b strings.Builder
	fmt.Fprintf(&b, "%s.Layout{Flat: %d", w.wire, len(t.Flat()))
	kind := wireKind(t)
	switch k := kind.(type) {
	case wit.String, *wit.List, *wit.Map:
		w.emitFields(&b, []wireField{{typ: wit.U32{}}, {typ: wit.U32{}, offset: 4}})
	case *wit.Record, *wit.Tuple:
		w.emitFields(&b, wireFields(k))
	case *wit.Option, *wit.Result, *wit.Variant:
		v := wit.Despecialize(kind).(*wit.Variant)
		w.emitFields(&b, []wireField{{typ: wit.Discriminant(len(v.Cases))}})
		fmt.Fprintf(&b, ", Cases: []%s.Field{", w.wire)
		for _, c := range v.Cases {
			if c.Type == nil {
				b.WriteString("{},")
				continue
			}
			fmt.Fprintf(&b, "{Offset: %d, Layout: &%s},", payloadOffset(v), w.layout(c.Type))
		}
		b.WriteString("}")
	case *wit.Flags:
		var fields []wireField
		if n := len(t.Flat()); n > 1 {
			for i := range n {
				fields = append(fields, wireField{typ: wit.U32{}, offset: uintptr(i) * 4})
			}
			w.emitFields(&b, fields)
		} else {
			fmt.Fprintf(&b, ", Scalar: %d", t.Size())
		}
	default:
		switch k.(type) {
		case wit.S8, wit.S16:
			fmt.Fprintf(&b, ", Scalar: %s.Signed | %d", w.wire, t.Size())
		default:
			fmt.Fprintf(&b, ", Scalar: %d", t.Size())
		}
	}
	b.WriteString("}")
	fmt.Fprintf(&w.declarations, "var %s = %s\n", name, b.String())
	return name
}

func (w *wasm32Generator) emitFields(b *strings.Builder, fields []wireField) {
	fmt.Fprintf(b, ", Fields: []%s.Field{", w.wire)
	for _, f := range fields {
		fmt.Fprintf(b, "{Offset: %d, Layout: &%s},", f.offset, w.layout(f.typ))
	}
	b.WriteString("}")
}
