package vacuum_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/vacuum"
)

func TestTimestampChecker_Expired(t *testing.T) {
	t.Parallel()

	data := []struct {
		name      string
		timestamp string
		exp       bool
	}{
		{
			name:      "expired",
			timestamp: "2025-01-13T00:14:59+09:00",
			exp:       true,
		},
		{
			name:      "not expired",
			timestamp: "2025-01-13T00:15:01+09:00",
		},
	}

	now, err := vacuum.ParseTime("2025-01-20T00:15:00+09:00")
	if err != nil {
		t.Fatal(err)
	}

	checker := vacuum.NewTimestampChecker(now, 7)

	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts, err := vacuum.ParseTime(tt.timestamp)
			if err != nil {
				t.Fatal(err)
			}

			a := checker.Expired(ts)
			if a != tt.exp {
				t.Fatalf("wanted %v, got %v", tt.exp, a)
			}
		})
	}
}
