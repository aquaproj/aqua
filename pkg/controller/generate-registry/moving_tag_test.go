package genrgst_test

import (
	"testing"

	genrgst "github.com/aquaproj/aqua/v2/pkg/controller/generate-registry"
)

// A moving tag names a place in a release history rather than a point in it. A caller
// that records what one points at records something that stops being true the next
// time the repository releases.
func TestIsMovingTag(t *testing.T) {
	t.Parallel()
	for _, d := range []struct {
		tag  string
		want bool
	}{
		{tag: "stable", want: true},
		{tag: "latest", want: true},
		{tag: "nightly", want: true},
		{tag: "v1.0.0", want: false},
		// Only the tag itself moves. One that merely contains the word is a release
		// like any other.
		{tag: "v1.0.0-stable", want: false},
		{tag: "stable-v1.0.0", want: false},
	} {
		t.Run(d.tag, func(t *testing.T) {
			t.Parallel()
			if got := genrgst.IsMovingTag(d.tag); got != d.want {
				t.Errorf("IsMovingTag(%q) is %v, want %v", d.tag, got, d.want)
			}
		})
	}
}
