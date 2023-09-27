package fuzzyfinder_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
)

func TestVersion_Preview(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		version *fuzzyfinder.Version
		w       int
		exp     string
	}{
		{
			name: "tag",
			version: &fuzzyfinder.Version{
				Version: "v1.0.0",
				Name:    "v1.0.0",
			},
			w:   100,
			exp: `v1.0.0`,
		},
		{
			name: "release",
			version: &fuzzyfinder.Version{
				Version: "v1.0.0",
				Name:    "Major",
				URL:     "https://github.com/suzuki-shunsuke/tfcmt/releases/v1.0.0",
				Description: `foo
bar`,
			},
			w: 100,
			exp: `v1.0.0 (Major)

https://github.com/suzuki-shunsuke/tfcmt/releases/v1.0.0
foo
bar`,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s := d.version.Preview(d.w)
			if s != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
