package installpackage

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func TestIsReservedCommand(t *testing.T) {
	t.Parallel()
	for cmd, want := range map[string]bool{
		"aqua":       true,
		"AQUA":       true,
		"aqua-proxy": true,
		"Aqua-Proxy": true,
		"gh":         false,
		"aquas":      false,
		"aqua-":      false,
	} {
		if got := isReservedCommand(cmd); got != want {
			t.Errorf("isReservedCommand(%q) = %v, want %v", cmd, got, want)
		}
	}
}

// $(aqua root-dir)/bin/aqua is aqua itself. A link there would replace it with a link to
// aqua-proxy, which refuses to run aqua, so the link is skipped and the binary is kept.
func TestInstaller_createCmdLink_reserved(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	binDir := filepath.Join(rootDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	aquaPath := filepath.Join(binDir, "aqua")
	if err := os.WriteFile(aquaPath, []byte("aqua"), 0o755); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	is := &Installer{
		rootDir:     rootDir,
		linker:      link.New(),
		realRuntime: &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
	}
	logger := slog.New(slog.DiscardHandler)
	if err := is.createCmdLink(logger, &registry.File{Name: "aqua"}, "aqua", ""); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Lstat(aquaPath)
	if err != nil {
		t.Fatal(err)
	}
	if !fi.Mode().IsRegular() {
		t.Errorf("bin/aqua must stay the binary, got mode %v", fi.Mode())
	}
	// Any other command is linked as before.
	if err := is.createCmdLink(logger, &registry.File{Name: "gh"}, "gh", ""); err != nil {
		t.Fatal(err)
	}
	fi, err = os.Lstat(filepath.Join(binDir, "gh"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("bin/gh must be a link, got mode %v", fi.Mode())
	}
}
