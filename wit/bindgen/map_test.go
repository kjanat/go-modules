//go:build !tinygo

package bindgen

import (
	"strings"
	"testing"

	"github.com/kjanat/go-modules/wit"
)

func TestGoMapBindings(t *testing.T) {
	const src = `package test:maps;
interface api {
	type table = map<string, u32>;
	echo: func(input: table) -> table;
	pairs: func(input: option<map<string, string>>);
}
world example {
	import api;
	export api;
}
`
	res, err := wit.DecodeWIT(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Go(res, GeneratedBy("test"))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			src, err := file.Bytes()
			if err != nil {
				t.Fatal(err)
			}
			b.Write(src)
		}
	}
	generated := b.String()
	for _, want := range []string{
		`"github.com/kjanat/go-modules/cm"`,
		"type Table cm.Map[string, uint32]",
		"func Echo(input Table) (result Table)",
		"input0, input1 := cm.LowerMap[cm.Tuple[string, uint32]](input)",
		"func wasmimport_Echo(input0 *cm.Tuple[string, uint32], input1 uint32, result *Table)",
		"input := cm.LiftMap[Table]((*cm.Tuple[string, uint32])(input0), (uint32)(input1))",
		"func Pairs(input cm.Option[cm.Map[string, string]])",
		"v1, v2 := cm.LowerMap[[2]string](*some)",
		"cm.LiftMap[cm.Map[string, string]]((*[2]string)(f1), (uint32)(f2))",
	} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated Go is missing %q", want)
		}
	}
	const custom = "example.com/custom/cm"
	pkgs, err = Go(res, GeneratedBy("test"), CMPackage(custom))
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range pkgs {
		if file := pkg.Files["map_abi_wasip1.go"]; file != nil {
			if _, ok := file.Imports[custom+"/wasm32"]; !ok {
				t.Errorf("map binding imports = %v, want %s/wasm32", file.Imports, custom)
			}
			if _, ok := file.Imports[cmPackage+"/wasm32"]; ok {
				t.Errorf("map binding also imports default runtime %s/wasm32", cmPackage)
			}
			return
		}
	}
	t.Fatal("no standard Go map binding was generated")
}
