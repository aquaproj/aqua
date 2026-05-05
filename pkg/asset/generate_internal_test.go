package asset

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
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
			goos: osLinux,
			m1: map[string]string{
				// linux/amd64
				osLinux:   osLinuxCapitalized,
				archAmd64: archX86_64,
			},
			m2: map[string]string{
				// linux/arm64
				osLinux: osLinuxCapitalized,
			},
			m: map[string]string{
				// linux
				osLinux:   osLinuxCapitalized,
				archAmd64: archX86_64,
			},
			f: true,
		},
		{
			name: "conflicted",
			goos: osLinux,
			m1: map[string]string{
				// linux/amd64
				osLinux:   osLinuxCapitalized,
				archAmd64: archX86_64,
			},
			m2: nil,
			m:  nil,
			f:  false,
		},
	}
	for _, d := range data {
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
		overrides    []*registry.Override
		replacements map[string]string
		exp          []*registry.Override
	}{
		{
			name: "normal",
			overrides: []*registry.Override{
				{
					GOOS: osLinux,
					Replacements: map[string]string{
						osLinux: osLinuxCapitalized,
					},
				},
			},
			replacements: map[string]string{
				osLinux: osLinuxCapitalized,
			},
			exp: nil,
		},
		{
			name: "complicated",
			overrides: []*registry.Override{
				{
					GOOS: osLinux,
					Replacements: map[string]string{
						osLinux: osLinuxCapitalized,
					},
				},
				{
					GOOS: osDarwin,
					Replacements: map[string]string{
						archAmd64: archX86_64,
					},
				},
			},
			replacements: map[string]string{
				osLinux: osLinuxCapitalized,
			},
			exp: []*registry.Override{
				{
					GOOS: osDarwin,
					Replacements: map[string]string{
						archAmd64: archX86_64,
					},
				},
			},
		},
		{
			name: "complicated 2",
			overrides: []*registry.Override{
				{
					GOOS: osLinux,
					Replacements: map[string]string{
						osLinux:   osLinuxCapitalized,
						archAmd64: archX86_64,
					},
				},
				{
					GOOS: osDarwin,
					Replacements: map[string]string{
						archAmd64: archX86_64,
						osDarwin:  "MacOS",
					},
				},
				{
					GOOS:   osWindows,
					Format: formatZip,
					Replacements: map[string]string{
						archAmd64: archX86_64,
						osWindows: "Windows",
					},
				},
			},
			replacements: map[string]string{
				osLinux:   osLinuxCapitalized,
				osDarwin:  "MacOS",
				osWindows: "Windows",
				archAmd64: archX86_64,
			},
			exp: []*registry.Override{
				{
					GOOS:   osWindows,
					Format: formatZip,
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			m, overrides := normalizeOverridesByReplacements(&registry.PackageInfo{
				Overrides: d.overrides,
			})
			if diff := cmp.Diff(d.replacements, m); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(d.exp, overrides); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
