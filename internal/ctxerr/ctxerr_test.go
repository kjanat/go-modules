package ctxerr

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestJoinPreservesErrorIdentity(t *testing.T) {
	base := context.Background()
	diagnostic := errors.New("operation failed")
	if got := Join(base, diagnostic); got != diagnostic {
		t.Fatalf("live context changed the error: %v", got)
	}
	ctx, cancel := context.WithDeadline(base, time.Now().Add(-time.Second))
	defer cancel()
	if got := Join(ctx, nil); got != context.DeadlineExceeded {
		t.Fatalf("lost deadline on success: %v", got)
	}
	wrapped := fmt.Errorf("fetch: %w", context.DeadlineExceeded)
	if got := Join(ctx, wrapped); got != wrapped {
		t.Fatalf("already-matching error was wrapped again: %v", got)
	}
	if got := Join(ctx, diagnostic); !errors.Is(got, diagnostic) || !errors.Is(got, context.DeadlineExceeded) {
		t.Fatalf("lost diagnostic or deadline: %v", got)
	}
}
