//go:build linux

package runtime

import (
	"bytes"
	"context"
	"os"
	"os/exec"
)

const (
	libcMusl  = "musl"
	libcGlibc = "glibc"
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

// glibcLdFiles lists the paths of the glibc dynamic linker on the architectures
// aqua runs on. Debian's multiarch layout keeps the loader under
// /lib/<triplet>/, and /lib64 or /lib holds the canonical name as a symlink,
// which a stat follows.
var glibcLdFiles = []string{ //nolint:gochecknoglobals
	"/lib64/ld-linux-x86-64.so.2",
	"/lib/ld-linux-x86-64.so.2",
	"/lib/ld-linux-aarch64.so.1",
	"/lib/ld-linux-armhf.so.3",
}

// detectLibC returns the libc implementation in use on the current Linux system.
// It returns "musl", "glibc", or "" when detection is not possible.
//
// Detection is tiered:
//  1. Stat the well-known paths of each implementation's dynamic linker. No
//     subprocess, and it works on distroless or other images that do not ship
//     ldd. This answers on every mainstream distribution.
//  2. Run `ldd --version` and inspect the combined stdout/stderr. musl's ldd
//     exits non-zero but writes "musl libc..." to stderr; glibc's ldd writes
//     "ldd (GNU libc)..." to stdout. The exit code is intentionally ignored.
//
// The order matters for what it costs. aqua asks for the libc on every
// invocation, including every command run through its proxy, and a subprocess
// is worth thousands of stats: `ldd` is a shell script that starts the dynamic
// loader. Recording the answer somewhere would avoid the subprocess too, but a
// recorded answer travels: a home directory mounted into an Alpine container,
// or a CI cache restored onto another image, would be read by an environment
// the answer wasn't for. A stat asks the filesystem aqua is actually running
// on, which cannot be stale.
//
// When neither method yields a positive signal (e.g. ldd is missing and no
// loader is where it is expected), an empty string is returned so
// libc-constrained overrides do not match.
func detectLibC(ctx context.Context) string {
	// musl first: a system carrying both, such as Alpine with gcompat, runs musl
	// binaries, and the glibc loader is there for the things that need it.
	for _, p := range muslLdFiles {
		if _, err := os.Stat(p); err == nil {
			return libcMusl
		}
	}
	for _, p := range glibcLdFiles {
		if _, err := os.Stat(p); err == nil {
			return libcGlibc
		}
	}
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
	return libcGlibc
}
