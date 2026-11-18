// Package ctxerr preserves cancellation alongside operation errors.
package ctxerr

import (
	"context"
	"errors"
)

// Join returns err with the context error, if any, without duplicating a cause
// the operation already reported.
func Join(ctx context.Context, err error) error {
	cause := ctx.Err()
	if err == nil {
		return cause
	}
	if cause == nil || errors.Is(err, cause) {
		return err
	}
	return errors.Join(err, cause)
}
