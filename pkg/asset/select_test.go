package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/google/go-cmp/cmp"
)

type picked struct {
	Name     string
	Libc     string
	Emulated bool
}

// summarize maps each selected asset to "<os>/<arch>" plus its libc, so a test can say
// which environment got what without restating how the name was parsed.
func summarize(sel *asset.Selection) map[string]picked {
	out := map[string]picked{}
	for _, a := range sel.Assets {
		k := a.Env()
		if a.Libc != "" {
			k += "/" + a.Libc
		}
		out[k] = picked{Name: a.Name, Libc: a.Libc, Emulated: a.Emulated}
	}
	return out
}

func TestSelect(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		pkgName string
		version string
		assets  []string
		want    map[string]picked
		ties    []string
		unnamed []string
	}{
		{
			name:    "goreleaser layout with checksums and signatures",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets: []string{
				"checksums.txt", "checksums.txt.sig", "checksums.txt.pem",
				"foo_1.0.0_darwin_amd64.tar.gz", "foo_1.0.0_darwin_arm64.tar.gz",
				"foo_1.0.0_linux_amd64.tar.gz", "foo_1.0.0_linux_arm64.tar.gz",
				"foo_1.0.0_linux_amd64.tar.gz.sigstore.json",
				"foo_1.0.0_windows_amd64.zip", "foo_1.0.0_windows_arm64.zip",
				"foo_1.0.0_linux_386.tar.gz", "foo_1.0.0_linux_amd64.deb",
			},
			want: map[string]picked{
				"darwin/amd64":  {Name: "foo_1.0.0_darwin_amd64.tar.gz"},
				"darwin/arm64":  {Name: "foo_1.0.0_darwin_arm64.tar.gz"},
				"linux/amd64":   {Name: "foo_1.0.0_linux_amd64.tar.gz"},
				"linux/arm64":   {Name: "foo_1.0.0_linux_arm64.tar.gz"},
				"windows/amd64": {Name: "foo_1.0.0_windows_amd64.zip"},
				"windows/arm64": {Name: "foo_1.0.0_windows_arm64.zip"},
			},
		},
		{
			name:    "rust target triples give a glibc and a musl build",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets: []string{
				"foo-x86_64-unknown-linux-gnu.tar.gz", "foo-x86_64-unknown-linux-musl.tar.gz",
				"foo-aarch64-unknown-linux-musl.tar.gz",
				"foo-x86_64-apple-darwin.tar.gz", "foo-aarch64-apple-darwin.tar.gz",
				"foo-x86_64-pc-windows-msvc.zip", "foo-x86_64-pc-windows-gnu.zip",
			},
			want: map[string]picked{
				"darwin/amd64":      {Name: "foo-x86_64-apple-darwin.tar.gz"},
				"darwin/arm64":      {Name: "foo-aarch64-apple-darwin.tar.gz"},
				"linux/amd64/glibc": {Name: "foo-x86_64-unknown-linux-gnu.tar.gz", Libc: "glibc"},
				"linux/amd64/musl":  {Name: "foo-x86_64-unknown-linux-musl.tar.gz", Libc: "musl"},
				"linux/arm64/musl":  {Name: "foo-aarch64-unknown-linux-musl.tar.gz", Libc: "musl"},
				"windows/amd64":     {Name: "foo-x86_64-pc-windows-msvc.zip"},
				"windows/arm64":     {Name: "foo-x86_64-pc-windows-msvc.zip", Emulated: true},
			},
		},
		{
			name:    "a musl build beside one saying no libc",
			pkgName: "UsefulSoftwareCo/executor",
			version: "v1.6.10",
			assets:  []string{"executor-linux-x64.tar.gz", "executor-linux-x64-musl.tar.gz"},
			want: map[string]picked{
				"linux/amd64":      {Name: "executor-linux-x64.tar.gz"},
				"linux/amd64/musl": {Name: "executor-linux-x64-musl.tar.gz", Libc: "musl"},
			},
		},
		{
			name:    "a universal darwin build is for both architectures",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets:  []string{"foo_1.0.0_darwin_universal.tar.gz", "foo_1.0.0_linux_amd64.tar.gz"},
			want: map[string]picked{
				"darwin/amd64": {Name: "foo_1.0.0_darwin_universal.tar.gz"},
				"darwin/arm64": {Name: "foo_1.0.0_darwin_universal.tar.gz"},
				"linux/amd64":  {Name: "foo_1.0.0_linux_amd64.tar.gz"},
			},
		},
		{
			name:    "an architecture's own build beats the universal one",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets:  []string{"foo_darwin_all.tar.gz", "foo_darwin_arm64.tar.gz"},
			want: map[string]picked{
				"darwin/amd64": {Name: "foo_darwin_all.tar.gz"},
				"darwin/arm64": {Name: "foo_darwin_arm64.tar.gz"},
			},
		},
		{
			name:    "Rosetta 2",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets:  []string{"foo-darwin-amd64.tar.gz"},
			want: map[string]picked{
				"darwin/amd64": {Name: "foo-darwin-amd64.tar.gz"},
				"darwin/arm64": {Name: "foo-darwin-amd64.tar.gz", Emulated: true},
			},
		},
		{
			name:    "a Windows build naming no architecture isn't emulated",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets:  []string{"foo.exe"},
			want: map[string]picked{
				"windows/amd64": {Name: "foo.exe"},
			},
		},
		{
			name:    "an archive beats a bare executable, and a dmg loses to a tarball",
			pkgName: "foo/foo",
			version: "v1.0.0",
			assets: []string{
				"foo-windows-amd64.exe", "foo-windows-amd64.zip",
				"foo-darwin-arm64.dmg", "foo-darwin-arm64.tar.gz",
			},
			want: map[string]picked{
				"darwin/arm64":  {Name: "foo-darwin-arm64.tar.gz"},
				"windows/amd64": {Name: "foo-windows-amd64.zip"},
				"windows/arm64": {Name: "foo-windows-amd64.zip", Emulated: true},
			},
		},
		{
			name:    "a binary beats a file named after it",
			pkgName: "regclient/regclient",
			version: "v0.11.6",
			assets:  []string{"regctl-linux-amd64", "regctl-linux-amd64.provenance"},
			want: map[string]picked{
				"linux/amd64": {Name: "regctl-linux-amd64"},
			},
		},
		{
			name:    "a build for particular hardware ties with the ordinary one",
			pkgName: "ollama/ollama",
			version: "v0.5.0",
			assets: []string{
				"ollama-linux-amd64.tgz", "ollama-linux-amd64-rocm.tgz",
				"ollama-linux-arm64.tgz",
			},
			want: map[string]picked{
				"linux/arm64": {Name: "ollama-linux-arm64.tgz"},
			},
			ties: []string{"linux/amd64: ollama-linux-amd64-rocm.tgz, ollama-linux-amd64.tgz"},
		},
		{
			name:    "an asset naming no OS is left for a rule",
			pkgName: "koalaman/shellcheck",
			version: "v0.11.0",
			assets: []string{
				"shellcheck-v0.11.0.linux.x86_64.tar.xz", "shellcheck-v0.11.0.darwin.aarch64.tar.xz",
				"shellcheck-v0.11.0.zip",
			},
			want: map[string]picked{
				"darwin/arm64": {Name: "shellcheck-v0.11.0.darwin.aarch64.tar.xz"},
				"linux/amd64":  {Name: "shellcheck-v0.11.0.linux.x86_64.tar.xz"},
			},
			unnamed: []string{"shellcheck-v0.11.0.zip"},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			sel := asset.Select(d.pkgName, d.version, d.assets, nil)
			if diff := cmp.Diff(d.want, summarize(sel)); diff != "" {
				t.Errorf("assets (-want +got):\n%s", diff)
			}
			var ties []string
			for _, tie := range sel.Ties {
				ties = append(ties, tie.String())
			}
			if diff := cmp.Diff(d.ties, ties); diff != "" {
				t.Errorf("ties (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(d.unnamed, sel.Unnamed); diff != "" {
				t.Errorf("unnamed (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSelectAs(t *testing.T) {
	t.Parallel()
	sel := asset.SelectAs("windows", []string{"amd64", "arm64"}, []string{"shellcheck-v0.11.0.zip"}, "v0.11.0", nil)
	want := map[string]picked{
		"windows/amd64": {Name: "shellcheck-v0.11.0.zip"},
		"windows/arm64": {Name: "shellcheck-v0.11.0.zip"},
	}
	if diff := cmp.Diff(want, summarize(sel)); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}
