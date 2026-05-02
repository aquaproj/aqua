//go:build linux

package runtime

import (
	"testing"
)

func TestDetectLibC(t *testing.T) {
	t.Parallel()
	got := detectLibC(t.Context())
	switch got {
	case "musl", "glibc", "":
	default:
		t.Fatalf("unexpected libc value: %q (want one of \"musl\", \"glibc\", \"\")", got)
	}
}
