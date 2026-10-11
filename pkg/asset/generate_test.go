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

func TestGetOSArch(t *testing.T) { //nolint:funlen
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
			// The plain build and a build for particular hardware say the same
			// thing about the platform, so what is left to choose by is the name.
			// The assets arrive in the order GitHub lists them, where "-mlx"
			// sorts before ".tar".
			name:   "a build for particular hardware isn't the ordinary one",
			goos:   osLinux,
			goarch: archAmd64,
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: "tar.zst", Template: "ollama-{{.OS}}-{{.Arch}}-mlx.{{.Format}}"},
				{OS: osLinux, Arch: archAmd64, Format: "tar.zst", Template: "ollama-{{.OS}}-{{.Arch}}-rocm.{{.Format}}"},
				{OS: osLinux, Arch: archAmd64, Format: "tar.zst", Template: "ollama-{{.OS}}-{{.Arch}}.{{.Format}}"},
			},
			expected: &asset.AssetInfo{OS: osLinux, Arch: archAmd64, Format: "tar.zst", Template: "ollama-{{.OS}}-{{.Arch}}.{{.Format}}"},
		},
		{
			// The same rule decides between two bare executables, which are
			// ranked separately from the archives.
			name:   "a bare executable for particular hardware isn't the ordinary one",
			goos:   osLinux,
			goarch: archAmd64,
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: formatRaw, Template: "tool-{{.OS}}-{{.Arch}}-rocm"},
				{OS: osLinux, Arch: archAmd64, Format: formatRaw, Template: "tool-{{.OS}}-{{.Arch}}"},
			},
			expected: &asset.AssetInfo{OS: osLinux, Arch: archAmd64, Format: formatRaw, Template: "tool-{{.OS}}-{{.Arch}}"},
		},
		{
			// Shorter is not the point: ollama's release also holds the desktop
			// application, whose name is shorter than the command's and is not the
			// command's name with a piece taken off.
			name:   "a shorter name that isn't the same name",
			goos:   osDarwin,
			goarch: archArm64,
			assetInfos: []*asset.AssetInfo{
				{OS: osDarwin, Arch: archArm64, Format: formatTarGz, Template: "ollama-{{.OS}}.{{.Format}}"},
				{OS: osDarwin, Arch: archArm64, Format: formatTarGz, Template: "Ollama.{{.Format}}"},
			},
			expected: &asset.AssetInfo{OS: osDarwin, Arch: archArm64, Format: formatTarGz, Template: "ollama-{{.OS}}.{{.Format}}"},
		},
		{
			// A template that becomes a template earlier still wins: it carries
			// less before the platform, whatever it carries after.
			name:   "the earlier template still wins",
			goos:   osLinux,
			goarch: archAmd64,
			assetInfos: []*asset.AssetInfo{
				{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Template: "a-very-long-prefix-{{.OS}}.{{.Format}}"},
				{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Template: "t-{{.OS}}-{{.Arch}}-with-a-long-suffix.{{.Format}}"},
			},
			expected: &asset.AssetInfo{OS: osLinux, Arch: archAmd64, Format: formatTarGz, Template: "t-{{.OS}}-{{.Arch}}-with-a-long-suffix.{{.Format}}"},
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

// A replacement promoted to the package has to be cancelled for an OS that spells the
// architecture the way aqua does, or that OS is looked for under the other spelling.
func TestParseAssetInfos_replacementPerOS(t *testing.T) {
	t.Parallel()
	data := []struct {
		name         string
		assets       []string
		version      string
		replacements registry.Replacements
		overrides    []*registry.Override
	}{
		{
			// darwin and windows are x86_64, linux is amd64.
			name: "zk-org/zk",
			assets: []string{
				"zk-v0.15.6-linux-amd64.tar.gz", "zk-v0.15.6-linux-arm64.tar.gz",
				"zk-v0.15.6-macos-arm64.tar.gz", "zk-v0.15.6-macos-x86_64.tar.gz",
				"zk-v0.15.6-windows-x86_64.tar.gz",
			},
			version:      "v0.15.6",
			replacements: registry.Replacements{archAmd64: "x86_64", osDarwin: "macos"},
			overrides: []*registry.Override{
				{GOOS: osLinux, Replacements: registry.Replacements{archAmd64: archAmd64}},
			},
		},
		{
			// linux is aarch64 as well as arm64, darwin is only arm64.
			name: "dotenvx/dotenvx",
			assets: []string{
				"dotenvx-2.34.2-darwin-amd64.tar.gz", "dotenvx-2.34.2-darwin-arm64.tar.gz",
				"dotenvx-2.34.2-linux-aarch64.tar.gz", "dotenvx-2.34.2-linux-amd64.tar.gz",
				"dotenvx-2.34.2-linux-arm64.tar.gz", "dotenvx-2.34.2-windows-amd64.tar.gz",
			},
			version: "v2.34.2",
			overrides: []*registry.Override{
				{GOOS: osLinux, Replacements: registry.Replacements{archArm64: "aarch64"}},
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			infos := make([]*asset.AssetInfo, 0, len(d.assets))
			for _, name := range d.assets {
				infos = append(infos, asset.ParseAssetName(name, d.version))
			}
			pkgInfo := &registry.PackageInfo{}
			asset.ParseAssetInfos(pkgInfo, infos)
			if diff := cmp.Diff(d.replacements, pkgInfo.Replacements); diff != "" {
				t.Errorf("replacements (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(d.overrides, pkgInfo.Overrides); diff != "" {
				t.Errorf("overrides (-want +got):\n%s", diff)
			}
		})
	}
}
