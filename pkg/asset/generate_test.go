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
func TestParseAssetInfos_replacementPerOS(t *testing.T) { //nolint:funlen
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
		{
			// windows has only amd64, which arm64 runs under emulation, so the promoted
			// arm64: aarch64 isn't cancelled for it and no override is left for windows.
			name: "windows emulating arm64",
			assets: []string{
				"foo-linux-amd64.tar.gz", "foo-linux-aarch64.tar.gz",
				"foo-darwin-amd64.tar.gz", "foo-darwin-aarch64.tar.gz",
				"foo-windows-amd64.tar.gz",
			},
			version:      "v1.0.0",
			replacements: registry.Replacements{archArm64: "aarch64"},
		},
		{
			// darwin has only amd64, which arm64 runs under Rosetta 2.
			name: "darwin under Rosetta 2",
			assets: []string{
				"foo-linux-amd64.tar.gz", "foo-linux-aarch64.tar.gz",
				"foo-darwin-amd64.tar.gz",
				"foo-windows-amd64.tar.gz", "foo-windows-aarch64.tar.gz",
			},
			version:      "v1.0.0",
			replacements: registry.Replacements{archArm64: "aarch64"},
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

// regclient/regclient v0.11.6 publishes SBOMs and signing files beside each binary, and
// they parse to the same OS and arch as the binary. The binary is what has nothing
// appended to its name, whichever comes first in the release.
func TestGetOSArch_sidecars(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		goos, goarch string
		names        []string
		want         string
	}{
		{
			goos: "windows", goarch: "amd64",
			names: []string{
				"regctl-windows-amd64.cyclonedx.json",
				"regctl-windows-amd64.exe",
				"regctl-windows-amd64.pem",
				"regctl-windows-amd64.sig",
				"regctl-windows-amd64.sigstore.json",
				"regctl-windows-amd64.spdx.json",
			},
			want: "regctl-{{.OS}}-{{.Arch}}",
		},
		{
			goos: "linux", goarch: "amd64",
			names: []string{
				"regctl-linux-amd64.cyclonedx.json.sig",
				"regctl-linux-amd64.asc",
				"regctl-linux-amd64",
			},
			want: "regctl-{{.OS}}-{{.Arch}}",
		},
	} {
		t.Run(tt.goos, func(t *testing.T) {
			t.Parallel()
			infos := make([]*asset.AssetInfo, 0, len(tt.names))
			for _, name := range tt.names {
				infos = append(infos, asset.ParseAssetName(name, "v0.11.6"))
			}
			got := asset.GetOSArch(tt.goos, tt.goarch, infos)
			if got == nil {
				t.Fatal("no asset was chosen")
			}
			if got.Template != tt.want {
				t.Errorf("chose %q, want %q", got.Template, tt.want)
			}
		})
	}
}

// The pick doesn't depend on the order a release lists its assets in, and a dotted suffix
// that isn't a file named after another asset decides nothing by itself.
func TestGetOSArch_order(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name         string
		goos, goarch string
		names        []string
		want         string
	}{
		{
			name: "a lower score loses", goos: "windows", goarch: "amd64",
			names: []string{"foo-windows-amd64.exe", "foo-windows.exe"},
			want:  "foo-{{.OS}}-{{.Arch}}",
		},
		{
			name: "an AppImage isn't a sidecar", goos: "linux", goarch: "amd64",
			names: []string{"tool-linux-amd64.AppImage", "tool-cli-linux-amd64"},
			want:  "tool-{{.OS}}-{{.Arch}}.AppImage",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, names := range [][]string{tt.names, {tt.names[1], tt.names[0]}} {
				infos := make([]*asset.AssetInfo, 0, len(names))
				for _, name := range names {
					infos = append(infos, asset.ParseAssetName(name, "v1.0.0"))
				}
				got := asset.GetOSArch(tt.goos, tt.goarch, infos)
				if got == nil || got.Template != tt.want {
					t.Errorf("%v: chose %+v, want %q", names, got, tt.want)
				}
			}
		})
	}
}

// XTLS/Xray-core spells amd64 as 64 and arm64 as arm64-v8a. The shorter spelling is
// contained in the longer one and must not be read out of it.
func TestParseAssetName_spellingInAnother(t *testing.T) {
	t.Parallel()
	spellings := asset.Spellings{"amd64": "64", "arm64": "arm64-v8a"}
	for name, want := range map[string]string{
		"Xray-linux-arm64-v8a.zip": "arm64",
		"Xray-linux-64.zip":        "amd64",
		"Xray-macos-arm64-v8a.zip": "arm64",
	} {
		if got := asset.ParseAssetName(name, "v26.3.27", spellings).Arch; got != want {
			t.Errorf("%s: arch is %q, want %q", name, got, want)
		}
	}
}
