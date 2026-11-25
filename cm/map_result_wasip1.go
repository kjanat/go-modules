//go:build wasip1 && wasm && !tinygo

package cm

// MapResult represents a WIT result that contains a map. Both cases have
// typed storage so Go can trace any pointers in the active case.
type MapResult[Shape, OK, Err any] struct {
	mapResult[OK, Err]
}

type mapResult[OK, Err any] struct {
	isErr bool
	ok    OK
	err   Err
}

// SetOK sets the success case.
func (r *mapResult[OK, Err]) SetOK(ok OK) { *r = mapResult[OK, Err]{ok: ok} }

// SetErr sets the error case.
func (r *mapResult[OK, Err]) SetErr(err Err) { *r = mapResult[OK, Err]{isErr: true, err: err} }

// IsOK reports whether the result represents success.
func (r *mapResult[OK, Err]) IsOK() bool { return !r.isErr }

// IsErr reports whether the result represents an error.
func (r *mapResult[OK, Err]) IsErr() bool { return r.isErr }

// OK returns the success value when that case is active.
func (r *mapResult[OK, Err]) OK() *OK {
	if r.isErr {
		return nil
	}
	return &r.ok
}

// Err returns the error value when that case is active.
func (r *mapResult[OK, Err]) Err() *Err {
	if !r.isErr {
		return nil
	}
	return &r.err
}

// Result returns the active value and reports whether it represents an error.
func (r mapResult[OK, Err]) Result() (ok OK, err Err, isErr bool) {
	return r.ok, r.err, r.isErr
}
