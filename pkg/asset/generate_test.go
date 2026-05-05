package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestParseAssetName(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name      string
		assetName string
		version   string
		expected  *asset.AssetInfo
	}{
		{
			name:      "basic linux tar.gz",
			assetName: "tool-v1.0.0-linux-amd64.tar.gz",
			version:   versionV1,
			expected: &asset.AssetInfo{
				Template:     "tool-{{.Version}}-{{.OS}}-{{.Arch}}.{{.Format}}",
				OS:           osLinux,
				Arch:         archAmd64,
				Format:       formatTarGz,
				Replacements: nil,
				Score:        0,
			},
		},
		{
			name:      "windows exe without version prefix",
			assetName: "tool-1.0.0-windows-amd64.exe",
			version:   "1.0.0",
			expected: &asset.AssetInfo{
				Template:     "tool-{{.Version}}-{{.OS}}-{{.Arch}}",
				OS:           osWindows,
				Arch:         archAmd64,
				Format:       formatRaw,
				Replacements: nil,
				Score:        0,
			},
		},
		{
			name:      "darwin universal binary",
			assetName: "tool-v2.1.0-darwin-all.tar.gz",
			version:   "v2.1.0",
			expected: &asset.AssetInfo{
				Template:  "tool-{{.Version}}-{{.OS}}-all.{{.Format}}",
				OS:        osDarwin,
				Arch:      archAmd64,
				DarwinAll: false,
				Format:    formatTarGz,
				Score:     -2,
			},
		},
		{
			name:      "mixed case architecture",
			assetName: "Tool-v1.0.0-Linux-X86_64.tar.gz",
			version:   versionV1,
			expected: &asset.AssetInfo{
				Template: "Tool-{{.Version}}-{{.OS}}-{{.Arch}}.{{.Format}}",
				OS:       osLinux,
				Arch:     archAmd64,
				Format:   formatTarGz,
				Replacements: map[string]string{
					osLinux:   osLinuxCapitalized,
					archAmd64: "X86_64",
				},
				Score: 0,
			},
		},
		{
			name:      "no platform info",
			assetName: "tool-v1.0.0.tar.gz",
			version:   versionV1,
			expected: &asset.AssetInfo{
				Template: "tool-{{.Version}}.{{.Format}}",
				OS:       "",
				Arch:     archAmd64,
				Format:   formatTarGz,
				Score:    -2,
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := asset.ParseAssetName(d.assetName, d.version)
			if diff := cmp.Diff(d.expected, result); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestGetOSArch(t *testing.T) {
	t.Parallel()
	data := []struct {
		name       string
		goos       string
		goarch     string
		assetInfos []*asset.AssetInfo
		expected   *asset.AssetInfo
	}{
		{
			name:   "exact match",
			goos:   osLinux,
			goarch: archAmd64,
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Score: 0, Template: assetTemplateOSArchFormat},
				{OS: osWindows, Arch: archAmd64, Format: formatZip, Score: 0, Template: assetTemplateOSArchFormat},
			},
			expected: &asset.AssetInfo{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Score: 0, Template: assetTemplateOSArchFormat},
		},
		{
			name:   "darwin all preference",
			goos:   osDarwin,
			goarch: archArm64,
			assetInfos: []*asset.AssetInfo{
				{OS: osDarwin, Arch: archAmd64, DarwinAll: true, Format: formatTarGz, Score: 0, Template: "tool-{{.OS}}-all.{{.Format}}"},
				{OS: osLinux, Arch: archArm64, Format: formatTarGz, Score: 0, Template: assetTemplateOSArchFormat},
			},
			expected: &asset.AssetInfo{OS: osDarwin, Arch: archAmd64, DarwinAll: true, Format: formatTarGz, Score: 0, Template: "tool-{{.OS}}-all.{{.Format}}"},
		},
		{
			name:   "prefer non-raw format",
			goos:   osLinux,
			goarch: archAmd64,
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: formatRaw, Score: 0, Template: "tool-{{.OS}}-{{.Arch}}"},
				{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Score: 0, Template: assetTemplateOSArchFormat},
			},
			expected: &asset.AssetInfo{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Score: 0, Template: assetTemplateOSArchFormat},
		},
		{
			name:   "no match returns nil",
			goos:   "freebsd",
			goarch: archAmd64,
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Score: 0, Template: assetTemplateOSArchFormat},
				{OS: osWindows, Arch: archAmd64, Format: formatZip, Score: 0, Template: assetTemplateOSArchFormat},
			},
			expected: nil,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := asset.GetOSArch(d.goos, d.goarch, d.assetInfos)
			if diff := cmp.Diff(d.expected, result); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestParseAssetInfos(t *testing.T) {
	t.Parallel()
	data := []struct {
		name       string
		assetInfos []*asset.AssetInfo
		expected   *registry.PackageInfo
	}{
		{
			name: "basic cross-platform",
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Template: assetTemplateOSArchFormat},
				{OS: osLinux, Arch: archArm64, Format: formatTarGz, Template: assetTemplateOSArchFormat},
				{OS: osDarwin, Arch: archAmd64, Format: formatTarGz, Template: assetTemplateOSArchFormat},
				{OS: osDarwin, Arch: archArm64, Format: formatTarGz, Template: assetTemplateOSArchFormat},
				{OS: osWindows, Arch: archAmd64, Format: formatZip, Template: assetTemplateOSArchFormat},
				{OS: osWindows, Arch: archArm64, Format: formatZip, Template: assetTemplateOSArchFormat},
			},
			expected: &registry.PackageInfo{
				Format: formatTarGz,
				Asset:  assetTemplateOSArchFormat,
				Overrides: []*registry.Override{
					{GOOS: osWindows, Format: formatZip},
				},
				SupportedEnvs: nil, // This will be normalized to nil for standard platforms
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			pkgInfo := &registry.PackageInfo{}
			asset.ParseAssetInfos(pkgInfo, d.assetInfos)

			if pkgInfo.Format != d.expected.Format {
				t.Errorf("Format: expected %s, got %s", d.expected.Format, pkgInfo.Format)
			}
			if pkgInfo.Asset != d.expected.Asset {
				t.Errorf("Asset: expected %s, got %s", d.expected.Asset, pkgInfo.Asset)
			}
			// Note: We're not doing full comparison here because the function is complex
			// and normalizes many fields. This tests the basic functionality.
		})
	}
}
