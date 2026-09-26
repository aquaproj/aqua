package lockfile_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/google/go-cmp/cmp"
)

func TestLockFile_Find(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{
		Packages: []*lockfile.Package{
			{Name: "cli/cli", Version: "v2.1.0", OS: "darwin", Arch: "arm64", Asset: "darwin-arm64"},
			{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Asset: "linux-amd64"},
			{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Asset: "linux-amd64-musl", Variants: map[string]string{"libc": "musl"}},
			{Name: "cli/cli", Version: "v2.0.0", OS: "darwin", Arch: "arm64", Asset: "old"},
			{Name: "cli/cli", Version: "v2.1.0", OS: "linux", Arch: "amd64", Asset: "unknown-variant", Variants: map[string]string{"distro": "alpine"}},
		},
	}
	tests := []struct {
		name string
		rt   *runtime.Runtime
		want string
	}{
		{
			name: "os and arch",
			rt:   &runtime.Runtime{GOOS: "darwin", GOARCH: "arm64"},
			want: "darwin-arm64",
		},
		{
			// The entry constraining libc wins over the one constraining nothing.
			name: "the most specific variant",
			rt:   &runtime.Runtime{GOOS: "linux", GOARCH: "amd64", LibC: "musl"},
			want: "linux-amd64-musl",
		},
		{
			// The unconstrained entry is the fallback, not something the musl entry
			// shadows.
			name: "a variant that doesn't apply",
			rt:   &runtime.Runtime{GOOS: "linux", GOARCH: "amd64", LibC: "glibc"},
			want: "linux-amd64",
		},
		{
			name: "no entry for the environment",
			rt:   &runtime.Runtime{GOOS: "windows", GOARCH: "amd64"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := lf.Find("cli/cli", "v2.1.0", tt.rt)
			if tt.want == "" {
				if got != nil {
					t.Fatalf("got %+v, want nothing", got)
				}
				return
			}
			if got == nil {
				t.Fatal("no entry was found")
			}
			if got.Asset != tt.want {
				t.Errorf("got %q, want %q", got.Asset, tt.want)
			}
		})
	}
}

// A variant key aqua doesn't evaluate can't be shown to hold, so the entry must not be
// installed on the strength of the keys that do.
func TestLockFile_Find_unsupportedVariantKey(t *testing.T) {
	t.Parallel()
	lf := &lockfile.LockFile{
		Packages: []*lockfile.Package{
			{Name: "foo/foo", Version: "v1.0.0", OS: "linux", Arch: "amd64", Variants: map[string]string{"distro": "alpine"}},
		},
	}
	if got := lf.Find("foo/foo", "v1.0.0", &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}); got != nil {
		t.Errorf("got %+v, want nothing", got)
	}
}

func TestPackage_PackageInfo(t *testing.T) {
	t.Parallel()
	pkg := &lockfile.Package{
		Name:      "cli/cli",
		Version:   "v2.1.0",
		OS:        "darwin",
		Arch:      "arm64",
		Type:      "github_release",
		RepoOwner: "cli",
		RepoName:  "cli",
		Asset:     "gh_2.1.0_macOS_arm64.zip",
		Format:    "zip",
		Files:     []*lockfile.File{{Name: "gh", Src: "gh_2.1.0_macOS_arm64/bin/gh"}},
	}
	got := pkg.PackageInfo()
	if got.Asset != pkg.Asset {
		t.Errorf("Asset is %q, want %q", got.Asset, pkg.Asset)
	}
	if diff := cmp.Diff([]string{"gh"}, []string{got.Files[0].Name}); diff != "" {
		t.Errorf("the files are wrong (-want +got):\n%s", diff)
	}
	// The name is already final, so nothing may append to it.
	if got.GetAppendExt() {
		t.Error("append_ext must be off")
	}
	if got.CompleteWindowsExt == nil || *got.CompleteWindowsExt {
		t.Error("complete_windows_ext must be off")
	}
}

func TestPackage_NeedsChecksum(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"github_release": true,
		"github_content": true,
		"github_archive": true,
		"http":           true,
		"go_build":       true,
		// go_install and cargo build through another tool, and aqua downloads
		// nothing it could check.
		"go_install": false,
		"cargo":      false,
	}
	for typ, want := range tests {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			if got := (&lockfile.Package{Type: typ}).NeedsChecksum(); got != want {
				t.Errorf("NeedsChecksum is %v, want %v", got, want)
			}
		})
	}
}
