package cosign

import (
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
