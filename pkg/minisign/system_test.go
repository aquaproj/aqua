package minisign_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/minisign"
)

func TestLookSystemExe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable resolution on Windows depends on PATHEXT")
	}
	dir := t.TempDir()
	exe := filepath.Join(dir, "minisign")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	got, err := minisign.LookSystemExe()
	if err != nil {
		t.Fatal(err)
	}
	if got != exe {
		t.Fatalf("LookSystemExe() = %q, want %q", got, exe)
	}
}

func TestLookSystemExe_notFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := minisign.LookSystemExe(); err == nil {
		t.Fatal("an error must be returned when minisign isn't installed on the system")
	}
}
