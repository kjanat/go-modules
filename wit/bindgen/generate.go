package bindgen

import (
	"github.com/kjanat/go-modules/internal/go/gen"
	"github.com/kjanat/go-modules/wit"
)

// Go generates one or more Go packages from [wit.Resolve] res.
// It returns any error that occurs during code generation.
func Go(res *wit.Resolve, opts ...Option) ([]*gen.Package, error) {
	g, err := newGenerator(res, opts...)
	if err != nil {
		return nil, err
	}
	return g.generate()
}
