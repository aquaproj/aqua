//go:build linux

package runtime

import (
	"bytes"
	"os/exec"
)

const (
	libcMusl = "musl"
	libcGNU  = "gnu"
)

// detectLibC returns the libc implementation in use on the current Linux system.
// It returns "musl", "gnu", or "" when detection is not possible (e.g. ldd is not on PATH).
// On musl systems, `ldd --version` exits non-zero but writes "musl libc..." to stderr;
// on glibc systems, it succeeds and writes "ldd (GNU libc)..." to stdout.
// We capture both streams and inspect the combined output.
func detectLibC() string {
	cmd := exec.Command("ldd", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	_ = cmd.Run()
	if bytes.Contains(out.Bytes(), []byte(libcMusl)) {
		return libcMusl
	}
	if out.Len() == 0 {
		return ""
	}
	return libcGNU
}
