package registry_test

import (
	"testing"

	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/runtime"
)

func TestOverride_Match(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exp      bool
		override *registry.Override
		rt       *runtime.Runtime
	}{
		{
			title: "goos doesn't match",
			override: &registry.Override{
				GOOS: "linux",
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
		{
			title: "goarch doesn't match",
			override: &registry.Override{
				GOArch: "arm64",
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
		{
			title: "match",
			exp:   true,
			override: &registry.Override{
				GOOS: "darwin",
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if f := d.override.Match(d.rt); f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
