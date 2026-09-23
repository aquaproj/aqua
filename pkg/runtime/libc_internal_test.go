package runtime

import (
	"testing"
)

func TestDetectLibCOnce(t *testing.T) {
	t.Parallel()
	// The machine's libc doesn't change while a process runs, so every caller
	// gets the same answer and only the first one pays for it.
	first := detectLibCOnce(t.Context())
	if second := detectLibCOnce(t.Context()); second != first {
		t.Fatalf("detectLibCOnce returned %q and then %q", first, second)
	}
	switch first {
	case "musl", "glibc", "":
	default:
		t.Fatalf("unexpected libc value: %q", first)
	}
}
