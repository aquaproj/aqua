package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// TestPackage_RenameFile checks what an executable is called once it is installed,
// against the two spellings a files[].src arrives with: a registry's, which leaves the
// Windows extension to aqua, and a resolved one -- a lock file's, or
// aqua-registry-g2's -- which already has it because that is the installed name.
func TestPackage_RenameFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	windows := &runtime.Runtime{GOOS: "windows", GOARCH: "amd64"}
	data := []struct {
		title   string
		holds   string
		src     string
		exp     string
		rt      *runtime.Runtime
		wantErr bool
	}{
		{
			// The registry's spelling, and an archive that doesn't carry the
			// extension: aqua renames the file so the command can be run.
			title: "the archive holds the name the registry gives",
			holds: "bin/gh",
			src:   "bin/gh",
			exp:   "bin/gh.exe",
			rt:    windows,
		},
		{
			// A resolved src of the same archive. Nothing has renamed anything yet,
			// so the rename still has to happen.
			title: "a resolved src of an archive that doesn't carry the extension",
			holds: "tree-sitter-windows-x64",
			src:   "tree-sitter-windows-x64.exe",
			exp:   "tree-sitter-windows-x64.exe",
			rt:    windows,
		},
		{
			// The archive carries the extension, so there is nothing to rename.
			title: "the archive holds the windows executable",
			holds: "bin/gh.exe",
			src:   "bin/gh.exe",
			exp:   "bin/gh.exe",
			rt:    windows,
		},
		{
			// An extension of its own is the file's name, not one aqua added.
			title: "another extension",
			holds: "bin/aqua.sh",
			src:   "bin/aqua.sh",
			exp:   "bin/aqua.sh",
			rt:    windows,
		},
		{
			// Nothing of the name is there under either spelling.
			title:   "nothing of the name",
			holds:   "README.md",
			src:     "bin/gh.exe",
			rt:      windows,
			wantErr: true,
		},
		{
			// Only Windows needs an extension at all.
			title: "not windows",
			holds: "bin/gh",
			src:   "bin/gh",
			exp:   "bin/gh",
			rt:    &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgPath := t.TempDir()
			held := filepath.Join(pkgPath, filepath.FromSlash(d.holds))
			if err := os.MkdirAll(filepath.Dir(held), 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(held, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			pkg := &Package{
				PackageInfo: &registry.PackageInfo{
					Type:      PkgInfoTypeGitHubRelease,
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     "gh.tar.gz",
				},
				Package: &aqua.Package{Version: "v2.1.0"},
			}
			s, err := pkg.RenameFile(logger, pkgPath, &registry.File{Name: "gh", Src: d.src}, d.rt)
			if d.wantErr {
				if err == nil {
					t.Fatal("an error should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if s != d.exp {
				t.Fatalf("got %s, want %s", s, d.exp)
			}
		})
	}
}
