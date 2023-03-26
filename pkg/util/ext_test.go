package util_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/util"
)

func TestExt(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		ext     string
		s       string
		version string
	}{
		{
			name:    "with version",
			ext:     "",
			s:       "foo_1.0.0",
			version: "v1.0.0",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ext := util.Ext(d.s, d.version)
			if ext != d.ext {
				t.Fatalf("wanted %s, got %s", d.ext, ext)
			}
		})
	}
}
