//go:build !tinygo

package wasmtools

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kjanat/wazero"
	"github.com/kjanat/wazero/api"
	"github.com/kjanat/wazero/imports/wasi_snapshot_preview1"
	"github.com/kjanat/wazero/sys"
)

// _start loops until test.stop returns true. Only the test's cleanup sets
// that flag: a cancelled Run must exit through wazero, not a normal return.
const loopingGuest = "\x00asm\x01\x00\x00\x00" +
	"\x01\x08\x02\x60\x00\x01\x7f\x60\x00\x00" + // () -> i32, () -> ()
	"\x02\x0d\x01\x04test\x04stop\x00\x00" + // import test.stop
	"\x03\x02\x01\x01\x07\x0a\x01\x06_start\x00\x01" +
	"\x0a\x0c\x01\x0a\x00\x03\x40\x10\x00\x45\x0d\x00\x0b\x0b" // loop { br_if 0 (eqz (call 0)) }

func TestRunInterruptsActiveGuest(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		want := context.Canceled
		if deadline {
			want = context.DeadlineExceeded
		}
		t.Run(want.Error(), func(t *testing.T) {
			base := t.Context()
			r := newRuntime(base)
			defer r.Close(context.WithoutCancel(base))
			entered := make(chan struct{})
			signal := sync.OnceFunc(func() { close(entered) })
			var stop atomic.Bool
			_, err := r.NewHostModuleBuilder("test").NewFunctionBuilder().WithFunc(func() uint32 {
				signal()
				runtime.Gosched()
				if stop.Load() {
					return 1
				}
				return 0
			}).Export("stop").Instantiate(base)
			if err != nil {
				t.Fatal(err)
			}
			instance, err := newInstance(base, r, func() ([]byte, error) { return []byte(loopingGuest), nil })
			if err != nil {
				t.Fatal(err)
			}
			var ctx context.Context
			var cancel context.CancelFunc
			if deadline {
				ctx, cancel = context.WithTimeout(base, time.Second)
			} else {
				ctx, cancel = context.WithCancel(base)
			}
			defer cancel()
			done := make(chan struct{})
			var runErr error
			go func() {
				defer close(done)
				runErr = instance.Run(ctx, nil, nil, nil, nil)
			}()
			defer func() { stop.Store(true); <-done }()
			guard := time.NewTimer(5 * time.Second)
			defer guard.Stop()
			select {
			case <-entered:
				if err := ctx.Err(); err != nil {
					t.Fatalf("context expired before testing active execution: %v", err)
				}
			case <-done:
				t.Fatalf("guest returned before entering its loop: %v", runErr)
			case <-guard.C:
				t.Fatal("guest did not enter its loop")
			}
			if !deadline {
				cancel()
			}
			select {
			case <-done:
				var exit *sys.ExitError
				if !errors.As(runErr, &exit) || !errors.Is(exit, want) {
					t.Fatalf("guest exit = %v, want wazero exit for %v", runErr, want)
				}
			case <-guard.C:
				t.Fatalf("guest kept running after %v", want)
			}
		})
	}
}

type closeRecorder struct {
	wazero.Runtime
	ctx context.Context
	err error
}

func (r *closeRecorder) Close(ctx context.Context) error {
	r.ctx = ctx
	return errors.Join(r.Runtime.Close(ctx), r.err)
}

func TestNewClosesRuntimeOnFailure(t *testing.T) {
	type key struct{}
	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), key{}, "request"))
	defer cancel()
	badBinary, closeErr := errors.New("read embedded binary"), errors.New("close failed")
	r := &closeRecorder{Runtime: newRuntime(ctx), err: closeErr}
	defer r.Runtime.Close(context.Background())
	var wasi api.Module
	inst, err := newInstance(ctx, r, func() ([]byte, error) {
		wasi = r.Module(wasi_snapshot_preview1.ModuleName)
		cancel()
		return nil, badBinary
	})
	if inst != nil || !errors.Is(err, badBinary) || !errors.Is(err, closeErr) {
		t.Fatalf("constructor result = (%v, %v), want both errors", inst, err)
	}
	if wasi == nil || !wasi.IsClosed() || r.ctx == nil || r.ctx.Done() != nil || r.ctx.Value(key{}) != "request" {
		t.Fatal("constructor did not close its WASI module with the detached request context")
	}
}
