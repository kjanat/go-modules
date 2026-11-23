package bindgen

import (
	"fmt"

	"github.com/kjanat/go-modules/internal/go/gen"
	"github.com/kjanat/go-modules/wit"
)

// Go generates one or more Go packages from [wit.Resolve] res.
// It returns any error that occurs during code generation.
func Go(res *wit.Resolve, opts ...Option) ([]*gen.Package, error) {
	// Reject maps before generating any code. They appear in TypeDefs even
	// when their only use is anonymous or nested in another type.
	for _, td := range res.TypeDefs {
		if m, ok := td.Kind.(*wit.Map); ok {
			return nil, fmt.Errorf("Go binding generation does not support WIT map types: %s", m.WIT(nil, td.TypeName()))
		}
	}
	g, err := newGenerator(res, opts...)
	if err != nil {
		return nil, err
	}
	return g.generate()
}
