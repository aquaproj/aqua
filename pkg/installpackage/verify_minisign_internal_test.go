package installpackage

import (
	"log/slog"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func writeFakeMinisign(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "minisign"), []byte("#!/bin/sh\n"), 0o755); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	return dir
}

func Test_minisignVerifier_Enabled(t *testing.T) { //nolint:funlen
	enabled := true
	disabled := false
	// linux/amd64 is one of the environments aqua manages a minisign binary for,
	// while linux/arm64 is not.
	supportedRT := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	unsupportedRT := &runtime.Runtime{GOOS: "linux", GOARCH: "arm64"}

	data := []struct {
		name       string
		rt         *runtime.Runtime
		minisign   *registry.Minisign
		pathHasExe bool
		want       bool
	}{
		{
			name:     "disabled",
			rt:       supportedRT,
			minisign: &registry.Minisign{Enabled: &disabled},
			want:     false,
		},
		{
			name:     "supported environment",
			rt:       supportedRT,
			minisign: &registry.Minisign{Enabled: &enabled},
			want:     true,
		},
		{
			name:       "unsupported environment falls back to a system minisign",
			rt:         unsupportedRT,
			minisign:   &registry.Minisign{Enabled: &enabled},
			pathHasExe: true,
			want:       true,
		},
		{
			name:     "unsupported environment without a system minisign",
			rt:       unsupportedRT,
			minisign: &registry.Minisign{Enabled: &enabled},
			want:     false,
		},
	}

	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			if d.pathHasExe && goruntime.GOOS == "windows" {
				t.Skip("executable resolution on Windows depends on PATHEXT")
			}
			if d.pathHasExe {
				t.Setenv("PATH", writeFakeMinisign(t))
			} else {
				t.Setenv("PATH", t.TempDir())
			}
			s := &minisignVerifier{
				runtime:  d.rt,
				minisign: d.minisign,
			}
			got, err := s.Enabled(logger)
			if err != nil {
				t.Fatal(err)
			}
			if got != d.want {
				t.Fatalf("Enabled() = %v, want %v", got, d.want)
			}
		})
	}
}
