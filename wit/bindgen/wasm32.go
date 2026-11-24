package bindgen

import (
	"fmt"
	"strings"

	"github.com/kjanat/go-modules/internal/go/gen"
	"github.com/kjanat/go-modules/wit"
)

func usesMap(f *wit.Function) bool {
	for _, params := range [][]wit.Param{f.Params, f.Results} {
		for _, p := range params {
			if containsMap(p.Type) {
				return true
			}
		}
	}
	return false
}

func containsMap(t wit.Type) bool {
	td, ok := t.(*wit.TypeDef)
	if !ok {
		return false
	}
	switch k := td.Root().Kind.(type) {
	case *wit.Map:
		return true
	case *wit.List:
		return containsMap(k.Type)
	case *wit.Option:
		return containsMap(k.Type)
	case *wit.Result:
		return containsMap(k.OK) || containsMap(k.Err)
	case *wit.Variant:
		for _, c := range k.Cases {
			if containsMap(c.Type) {
				return true
			}
		}
	case *wit.Record:
		for _, f := range k.Fields {
			if containsMap(f.Type) {
				return true
			}
		}
	case *wit.Tuple:
		for _, field := range k.Types {
			if containsMap(field) {
				return true
			}
		}
	}
	return false
}

func (g *generator) mapABIFile(pkg *gen.Package, standard bool) *gen.File {
	name, build := "map_abi_legacy.go", "!wasip1 || !wasm || tinygo"
	if standard {
		name, build = "map_abi_wasip1.go", "wasip1 && wasm && !tinygo"
	}
	file := pkg.File(name)
	file.GoBuild = build
	file.GeneratedBy = g.opts.generatedBy
	return file
}

type wasm32Generator struct {
	g            *generator
	file         *gen.File
	wire         string
	codecs       map[wireCodecKey]string
	layouts      map[wit.Type]string
	declarations strings.Builder
}

type wireCodecKey struct {
	dir wit.Direction
	typ wit.Type
}

func (g *generator) defineStandardMapFunction(decl *funcDecl) {
	pkg := decl.goFunc.file.Package
	if g.mapABIs == nil {
		g.mapABIs = make(map[*gen.Package]*wasm32Generator)
	}
	w := g.mapABIs[pkg]
	if w == nil {
		file := g.mapABIFile(pkg, true)
		w = &wasm32Generator{
			g: g, file: file,
			wire:    file.Import(cmPackage + "/wasm32"),
			layouts: make(map[wit.Type]string),
			codecs:  make(map[wireCodecKey]string),
		}
		g.mapABIs[pkg] = w
	}
	w.define(decl)
	w.file.WriteString(w.declarations.String())
	w.declarations.Reset()
}

func flatCount(params []param) (n int) {
	for _, p := range params {
		n += len(p.typ.Flat())
	}
	return n
}

func memoryLayout(params []param) (offsets []uintptr, size, align uintptr) {
	align = 1
	for _, p := range params {
		size = wit.Align(size, p.typ.Align())
		offsets = append(offsets, size)
		size += p.typ.Size()
		align = max(align, p.typ.Align())
	}
	return offsets, wit.Align(size, align), align
}

func (w *wasm32Generator) define(decl *funcDecl) {
	d := &w.declarations
	f := decl.goFunc
	params, results := f.params, f.results
	indirectParams := flatCount(params) > wit.MaxFlatParams
	indirectResults := flatCount(results) > wit.MaxFlatResults
	rawName := w.file.DeclareName(decl.wasmFunc.name + "_memory32")
	rawParams := decl.wasmFunc.params
	rawResults := decl.wasmFunc.results
	var rawArgs, rawTypes []string
	for i, p := range rawParams {
		rawArgs = append(rawArgs, fmt.Sprintf("_cmArg%d", i))
		rawTypes = append(rawTypes, w.rawType(p))
	}
	var signature strings.Builder
	fmt.Fprintf(&signature, "func %s(", rawName)
	for i, name := range rawArgs {
		if i != 0 {
			signature.WriteString(", ")
		}
		fmt.Fprintf(&signature, "%s %s", name, rawTypes[i])
	}
	signature.WriteString(")")
	if len(rawResults) != 0 {
		fmt.Fprintf(&signature, " %s", w.rawType(rawResults[0]))
	}
	var body strings.Builder
	if decl.dir == wit.Imported {
		w.defineImport(decl, &body, rawName, rawTypes, rawResults, indirectParams, indirectResults)
		fmt.Fprintf(d, "//go:wasmimport %s\n%s\n\n%s\n", decl.linkerName, signature.String(), body.String())
		return
	}
	w.defineExport(decl, &body, rawArgs, rawParams, rawResults, indirectParams, indirectResults)
	fmt.Fprintf(d, "//go:wasmexport %s\n%s {\n%s}\n\n", decl.linkerName, signature.String(), body.String())
	if indirectResults {
		post := w.file.DeclareName(rawName + "_post")
		fmt.Fprintf(d, "//go:wasmexport cabi_post_%s\nfunc %s(addr uint32) { %s.ReleaseResult(addr) }\n\n", decl.linkerName, post, w.wire)
	}
}

func (w *wasm32Generator) rawType(p param) string {
	if isPointer(p.typ) {
		return "uint32"
	}
	return w.g.typeRep(w.file, p.dir, p.typ)
}

func (w *wasm32Generator) rawValue(typ, bits string) string {
	switch typ {
	case "float32":
		return w.file.Import("math") + ".Float32frombits(uint32(" + bits + "))"
	case "float64":
		return w.file.Import("math") + ".Float64frombits(" + bits + ")"
	default:
		return typ + "(" + bits + ")"
	}
}

func (w *wasm32Generator) rawBits(typ, value string) string {
	switch typ {
	case "float32":
		return "uint64(" + w.file.Import("math") + ".Float32bits(" + value + "))"
	case "float64":
		return w.file.Import("math") + ".Float64bits(" + value + ")"
	default:
		return "uint64(" + value + ")"
	}
}

func (w *wasm32Generator) defineImport(decl *funcDecl, b *strings.Builder, raw string, rawTypes []string, rawResults []param, indirectParams, indirectResults bool) {
	f := decl.goFunc
	b.WriteString(w.g.functionDocs(wit.Imported, decl.f, f.name))
	b.WriteString("func ")
	if f.isMethod() {
		fmt.Fprintf(b, "(%s %s) ", f.receiver.name, w.g.typeRep(w.file, f.receiver.dir, f.receiver.typ))
	}
	fmt.Fprintf(b, "%s%s {\n", f.name, w.g.functionSignature(w.file, f))
	arena := f.scope.DeclareName("_cmArena")
	fmt.Fprintf(b, "%s := new(%s.Arena)\n", arena, w.wire)
	arg := make([]string, 0, len(rawTypes))
	if indirectParams {
		addr := f.scope.DeclareName("_cmArgs")
		w.storeParams(b, addr, arena, f.params)
		arg = append(arg, addr)
	} else {
		flat := f.scope.DeclareName("_cmFlat")
		if n := flatCount(f.params); n != 0 {
			fmt.Fprintf(b, "var %s [%d]uint64\n", flat, n)
			w.flattenParams(b, f.scope, arena, flat, f.params)
			for i := range n {
				arg = append(arg, w.rawValue(rawTypes[i], fmt.Sprintf("%s[%d]", flat, i)))
			}
		}
	}
	result := f.scope.DeclareName("_cmResult")
	if indirectResults {
		_, size, align := memoryLayout(f.results)
		fmt.Fprintf(b, "%s := %s.Allocate(%d, %d)\n", result, arena, size, align)
		arg = append(arg, result)
	}
	call := raw + "(" + strings.Join(arg, ", ") + ")"
	if len(rawResults) == 0 {
		fmt.Fprintln(b, call)
	} else {
		fmt.Fprintln(b, result+" := "+call)
	}
	if indirectResults {
		w.loadParams(b, result, f.results)
	} else if len(rawResults) != 0 {
		flat := fmt.Sprintf("[]uint64{%s}", w.rawBits(w.rawType(rawResults[0]), result))
		w.unflattenParams(b, f.scope, arena, flat, f.results)
	}
	fmt.Fprintf(b, "%s.KeepAlive(%s)\nreturn\n}\n", w.file.Import("runtime"), arena)
}

func (w *wasm32Generator) defineExport(decl *funcDecl, b *strings.Builder, rawArgs []string, rawParams, rawResults []param, indirectParams, indirectResults bool) {
	f := decl.goFunc
	for _, p := range append(append([]param(nil), f.params...), f.results...) {
		fmt.Fprintf(b, "var %s %s\n", p.name, w.g.typeRep(w.file, p.dir, p.typ))
	}
	b.WriteString("_cmArena := new(" + w.wire + ".Arena)\n")
	if indirectParams {
		w.loadParams(b, rawArgs[0], f.params)
		fmt.Fprintf(b, "%s.ReleaseInput(%s)\n", w.wire, rawArgs[0])
	} else {
		for i, p := range rawParams {
			rawArgs[i] = w.rawBits(w.rawType(p), rawArgs[i])
		}
		if len(rawArgs) > 0 {
			fmt.Fprintf(b, "_cmFlat := []uint64{%s}\n", strings.Join(rawArgs, ", "))
			w.unflattenParams(b, f.scope, "_cmArena", "_cmFlat", f.params)
		}
	}
	fqName := w.file.GetName("Exports") + "." + f.name
	if t := decl.f.Type(); t != nil {
		scope := w.g.exportScopes[decl.owner]
		fqName = w.file.GetName("Exports") + "." + scope.GetName(GoName(t.TypeName(), true)) + "." + f.name
	}
	names := func(params []param) string {
		var names []string
		for _, p := range params {
			names = append(names, p.name)
		}
		return strings.Join(names, ", ")
	}
	if len(f.results) != 0 {
		fmt.Fprintf(b, "%s = ", names(f.results))
	}
	fmt.Fprintf(b, "%s(%s)\n", fqName, names(f.params))
	if indirectResults {
		w.storeParams(b, "_cmResult", "_cmArena", f.results)
		b.WriteString("_cmArena.RetainResult(_cmResult)\nreturn _cmResult\n")
	} else if len(rawResults) != 0 {
		b.WriteString("var _cmResult [1]uint64\n")
		w.flattenParams(b, f.scope, "_cmArena", "_cmResult", f.results)
		fmt.Fprintf(b, "return %s\n", w.rawValue(w.rawType(rawResults[0]), "_cmResult[0]"))
	} else {
		fmt.Fprintf(b, "%s.KeepAlive(_cmArena)\n", w.file.Import("runtime"))
	}
}

func (w *wasm32Generator) storeParams(b *strings.Builder, addr, arena string, params []param) {
	offsets, size, align := memoryLayout(params)
	fmt.Fprintf(b, "%s := %s.Allocate(%d, %d)\n", addr, arena, size, align)
	for i, p := range params {
		fmt.Fprintf(b, "%sStore(%s, %s+%d, %s)\n", w.codec(p.dir, p.typ), arena, addr, offsets[i], p.name)
	}
}

func (w *wasm32Generator) loadParams(b *strings.Builder, addr string, params []param) {
	offsets, _, _ := memoryLayout(params)
	for i, p := range params {
		fmt.Fprintf(b, "%s = %sLoad(%s+%d)\n", p.name, w.codec(p.dir, p.typ), addr, offsets[i])
	}
}

func (w *wasm32Generator) flattenParams(b *strings.Builder, scope gen.Scope, arena, flat string, params []param) {
	i := 0
	for _, p := range params {
		n := len(p.typ.Flat())
		if n == 0 {
			continue
		}
		addr := scope.DeclareName("_cmValue")
		fmt.Fprintf(b, "%s := %s.Allocate(%d, %d)\n", addr, arena, p.typ.Size(), p.typ.Align())
		fmt.Fprintf(b, "%sStore(%s, %s, %s)\n", w.codec(p.dir, p.typ), arena, addr, p.name)
		fmt.Fprintf(b, "%s.Flatten(%s, %s[%d:%d])\n", w.layout(p.typ), addr, flat, i, i+n)
		i += n
	}
}

func (w *wasm32Generator) unflattenParams(b *strings.Builder, scope gen.Scope, arena, flat string, params []param) {
	i := 0
	for _, p := range params {
		n := len(p.typ.Flat())
		if n == 0 {
			continue
		}
		addr := scope.DeclareName("_cmValue")
		fmt.Fprintf(b, "%s := %s.Allocate(%d, %d)\n", addr, arena, p.typ.Size(), p.typ.Align())
		fmt.Fprintf(b, "%s.Unflatten(%s, (%s)[%d:%d])\n", w.layout(p.typ), addr, flat, i, i+n)
		fmt.Fprintf(b, "%s = %sLoad(%s)\n", p.name, w.codec(p.dir, p.typ), addr)
		i += n
	}
}
