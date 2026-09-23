package cosign

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// The backoff doubles and then stops doubling, so raising the number of attempts
// can't turn a retry into an outage.
func TestWaitTime(t *testing.T) {
	t.Parallel()
	data := []struct {
		retryCount int
		atLeast    time.Duration
		atMost     time.Duration
	}{
		{retryCount: 1, atLeast: 2 * time.Second, atMost: 3 * time.Second},
		{retryCount: 4, atLeast: 16 * time.Second, atMost: 17 * time.Second},
		{retryCount: 20, atLeast: maxWait, atMost: maxWait + time.Second},
	}
	for _, d := range data {
		got := waitTime(d.retryCount)
		if got < d.atLeast || got > d.atMost {
			t.Errorf("waitTime(%d) = %s, want between %s and %s", d.retryCount, got, d.atLeast, d.atMost)
		}
	}
}

// A verifier that never ran hasn't rejected anything, and saying it did would answer
// a question nobody got to ask.
func TestRan(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		err  error
		exp  bool
	}{
		{
			name: "the executable isn't there",
			err:  &exec.Error{Name: "cosign", Err: exec.ErrNotFound},
		},
		{
			name: "the context ended",
			err:  context.DeadlineExceeded,
		},
		{
			// What cosign says when it doesn't accept the signature.
			name: "it ran and exited non-zero",
			err:  &exec.ExitError{ProcessState: &os.ProcessState{}},
			exp:  true,
		},
		{
			name: "wrapped",
			err:  fmt.Errorf("run cosign: %w", &exec.ExitError{ProcessState: &os.ProcessState{}}),
			exp:  true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := ran(d.err); got != d.exp {
				t.Fatalf("ran(%v) = %v, wanted %v", d.err, got, d.exp)
			}
		})
	}
}
