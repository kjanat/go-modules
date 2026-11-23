//go:build !tinygo && !wasip1 && !wasip2

package cm

import (
	"go/types"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestUpstreamAPI(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedTypes}
	pkgs, err := packages.Load(cfg, "go.bytecodealliance.org/cm", "github.com/kjanat/go-modules/cm")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("loaded %d packages; want 2", len(pkgs))
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 || pkg.IllTyped {
			t.Fatalf("loading %s: %v", pkg.PkgPath, pkg.Errors)
		}
	}
	up, fork := pkgs[0].Types, pkgs[1].Types
	if up.Path() != "go.bytecodealliance.org/cm" {
		up, fork = fork, up
	}
	unqualified := func(*types.Package) string { return "" }
	for _, name := range up.Scope().Names() {
		want := up.Scope().Lookup(name)
		if !want.Exported() {
			continue
		}
		got := fork.Scope().Lookup(name)
		if got == nil {
			t.Errorf("cm does not export %s", name)
			continue
		}
		switch want := want.(type) {
		case *types.TypeName:
			if _, ok := got.(*types.TypeName); !ok || !aliases(got.Type(), want.Type()) {
				t.Errorf("cm.%s is %s; want an alias of %s", name, got.Type(), want.Type())
			}
		case *types.Func:
			if g, w := types.ObjectString(got, unqualified), types.ObjectString(want, unqualified); g != w {
				t.Errorf("cm.%s is %s; want %s", name, g, w)
			}
		case *types.Const:
			if g, ok := got.(*types.Const); !ok || g.Val().ExactString() != want.Val().ExactString() {
				t.Errorf("cm.%s does not equal upstream %s", name, want.Val())
			}
		default:
			t.Errorf("cm.%s: unhandled upstream object %T", name, want)
		}
	}
}

func aliases(got, want types.Type) bool {
	if types.Identical(got, want) {
		return true
	}
	a, ok := got.(*types.Alias)
	if !ok {
		return false
	}
	n, ok := a.Rhs().(*types.Named)
	if !ok || n.Origin() != want {
		return false
	}
	args, params := n.TypeArgs(), a.TypeParams()
	if args.Len() != params.Len() {
		return false
	}
	for i := range args.Len() {
		if args.At(i) != params.At(i) {
			return false
		}
	}
	return true
}
