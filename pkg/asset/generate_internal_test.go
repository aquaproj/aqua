package asset

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
)

func Test_mergeReplacements(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		goos string
		m1   map[string]string
		m2   map[string]string
		m    map[string]string
		f    bool
	}{
		{
			name: "normal",
			goos: "linux",
			m1: map[string]string{
				// linux/amd64
				"linux": "Linux",
				"amd64": "x86_64",
			},
			m2: map[string]string{
				// linux/arm64
				"linux": "Linux",
			},
			m: map[string]string{
				// linux
				"linux": "Linux",
				"amd64": "x86_64",
			},
			f: true,
		},
		{
			name: "conflicted",
			goos: "linux",
			m1: map[string]string{
				// linux/amd64
				"linux": "Linux",
				"amd64": "x86_64",
			},
			m2: nil,
			m:  nil,
			f:  false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			m, f := mergeReplacements(d.goos, d.m1, d.m2)
			if f != d.f {
				t.Fatalf("wanted %v, got %v", d.f, f)
			}
			if diff := cmp.Diff(d.m, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_normalizeOverridesByReplacements(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		rts          []*runtime.Runtime
		overrides    []*registry.Override
		replacements map[string]string
		exp          []*registry.Override
	}{
		{
			name: "normal",
			rts: []*runtime.Runtime{
				{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
			},
			overrides: []*registry.Override{
				{
					GOOS: "linux",
					Replacements: map[string]string{
						"linux": "Linux",
					},
				},
			},
			replacements: map[string]string{
				"linux": "Linux",
			},
			exp: []*registry.Override{},
		},
		{
			name: "complicated",
			rts: []*runtime.Runtime{
				{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "amd64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
			},
			overrides: []*registry.Override{
				{
					GOOS: "linux",
					Replacements: map[string]string{
						"linux": "Linux",
					},
				},
				{
					GOOS: "darwin",
					Replacements: map[string]string{
						"amd64": "x86_64",
					},
				},
			},
			replacements: map[string]string{
				"linux": "Linux",
			},
			exp: []*registry.Override{
				{
					GOOS: "darwin",
					Replacements: map[string]string{
						"amd64": "x86_64",
					},
				},
			},
		},
		{
			name: "complicated 2",
			rts: []*runtime.Runtime{
				{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
				{
					GOOS:   "linux",
					GOARCH: "arm64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "amd64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
				{
					GOOS:   "windows",
					GOARCH: "amd64",
				},
			},
			overrides: []*registry.Override{
				{
					GOOS: "linux",
					Replacements: map[string]string{
						"linux": "Linux",
						"amd64": "x86_64",
					},
				},
				{
					GOOS: "darwin",
					Replacements: map[string]string{
						"amd64":  "x86_64",
						"darwin": "MacOS",
					},
				},
				{
					GOOS:   "windows",
					Format: "zip",
					Replacements: map[string]string{
						"amd64":   "x86_64",
						"windows": "Windows",
					},
				},
			},
			replacements: map[string]string{
				"linux":   "Linux",
				"darwin":  "MacOS",
				"windows": "Windows",
				"amd64":   "x86_64",
			},
			exp: []*registry.Override{
				{
					GOOS:   "windows",
					Format: "zip",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			m, overrides := normalizeOverridesByReplacements(d.rts, d.overrides)
			if diff := cmp.Diff(d.replacements, m); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(d.exp, overrides); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_normalizeReplacementsInOverrides(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name                string
		rts                 []*runtime.Runtime
		overrides           []*registry.Override
		defaultReplacements map[string]string
		exp                 []*registry.Override
	}{
		{
			name: "normal",
			rts: []*runtime.Runtime{
				{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
				{
					GOOS:   "linux",
					GOARCH: "arm64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "amd64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
				{
					GOOS:   "windows",
					GOARCH: "amd64",
				},
			},
			overrides: []*registry.Override{
				{
					GOOS: "linux",
					Replacements: map[string]string{
						"linux": "Linux",
						"amd64": "x86_64",
					},
				},
				{
					GOOS: "darwin",
					Replacements: map[string]string{
						"amd64":  "x86_64",
						"darwin": "MacOS",
					},
				},
				{
					GOOS:   "windows",
					Format: "zip",
					Replacements: map[string]string{
						"amd64":   "x86_64",
						"windows": "Windows",
					},
				},
			},
			defaultReplacements: nil,
			exp: []*registry.Override{
				{
					GOOS:   "linux",
					GOArch: "amd64",
					Replacements: map[string]string{
						"linux": "Linux",
						"amd64": "x86_64",
					},
				},
				{
					GOOS:   "linux",
					GOArch: "arm64",
					Replacements: map[string]string{
						"linux": "Linux",
					},
				},
				{
					GOOS:   "darwin",
					GOArch: "amd64",
					Replacements: map[string]string{
						"amd64":  "x86_64",
						"darwin": "MacOS",
					},
				},
				{
					GOOS:   "darwin",
					GOArch: "arm64",
					Replacements: map[string]string{
						"darwin": "MacOS",
					},
				},
				{
					GOOS:   "windows",
					GOArch: "amd64",
					Replacements: map[string]string{
						"amd64":   "x86_64",
						"windows": "Windows",
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			overrides := normalizeReplacementsInOverrides(d.rts, d.defaultReplacements, d.overrides)
			if diff := cmp.Diff(d.exp, overrides); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
