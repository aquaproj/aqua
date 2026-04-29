//go:build linux

package runtime

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"
)

const (
	libcMusl = "musl"
	libcGNU  = "gnu"
	// lddTimeout caps the wall-clock cost of `ldd --version`. Under normal
	// conditions the command returns in milliseconds; the timeout is a safety
	// net so a hung process can never block aqua's startup.
	lddTimeout = 5 * time.Second
)

// muslLdFiles lists well-known paths of the musl dynamic linker / libc alias
// on common architectures. Mirrors the checks in the official Claude Code
// install script (libc.musl-*.so.1) and adds the upstream-canonical names
// (ld-musl-*.so.1) so detection works on minimal images that do not ship ldd.
var muslLdFiles = []string{ //nolint:gochecknoglobals
	"/lib/ld-musl-x86_64.so.1",
	"/lib/ld-musl-aarch64.so.1",
	"/lib/libc.musl-x86_64.so.1",
	"/lib/libc.musl-aarch64.so.1",
}

// detectLibC returns the libc implementation in use on the current Linux system.
// It returns "musl", "gnu", or "" when detection is not possible.
//
// Detection is tiered:
//  1. Stat well-known musl ld files. No subprocess and works on distroless or
//     other images that do not ship ldd.
//  2. Run `ldd --version` and inspect the combined stdout/stderr. musl's ldd
//     exits non-zero but writes "musl libc..." to stderr; glibc's ldd writes
//     "ldd (GNU libc)..." to stdout. The exit code is intentionally ignored.
//
// When neither method yields a positive signal (e.g. ldd is missing and no
// musl files are present), an empty string is returned so libc-constrained
// overrides do not match.
func detectLibC() string {
	for _, p := range muslLdFiles {
		if _, err := os.Stat(p); err == nil {
			return libcMusl
		}
	}
	// detectLibC runs at startup and is not part of any cancellable user
	// operation, so it deliberately uses a fresh context with its own timeout
	// rather than propagating one from the caller.
	ctx, cancel := context.WithTimeout(context.Background(), lddTimeout) //nolint:contextcheck
	defer cancel()
	cmd := exec.CommandContext(ctx, "ldd", "--version")
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
