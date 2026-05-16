//go:build !linux

package runtime

import "context"

// detectLibC returns an empty string on non-Linux platforms because libc
// variants are only meaningful on Linux. An empty value will not match any
// libc variant constraint in registry overrides.
func detectLibC(_ context.Context) string {
	return ""
}
