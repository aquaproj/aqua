package runtime

import (
	"fmt"
	"os"
	"runtime"
)

const (
	amd64   = "amd64"
	arm64   = "arm64"
	darwin  = "darwin"
	linux   = "linux"
	windows = "windows"
)

type Runtime struct {
	GOOS   string
	GOARCH string
}

func New() *Runtime {
	return &Runtime{
		GOOS:   goos(),
		GOARCH: goarch(),
	}
}

func NewR() *Runtime {
	return &Runtime{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	}
}

func (rt *Runtime) IsWindows() bool {
	return rt.GOOS == windows
}

func (rt *Runtime) Env() string {
	return fmt.Sprintf("%s/%s", rt.GOOS, rt.GOARCH)
}

func (rt *Runtime) Arch(rosetta2, windowsARMEmulation bool) string {
	if rt.GOARCH == amd64 {
		return amd64
	}

	if rt.GOOS == darwin && rosetta2 {
		return amd64
	}

	if rt.IsWindows() && windowsARMEmulation {
		return amd64
	}

	return rt.GOARCH
}

func goos() string {
	if s := os.Getenv("AQUA_GOOS"); s != "" {
		return s
	}

	return runtime.GOOS
}

func goarch() string {
	if s := os.Getenv("AQUA_GOARCH"); s != "" {
		return s
	}

	return runtime.GOARCH
}

func GOOSList() []string {
	return []string{darwin, linux, windows}
}

func GOOSMap() map[string]struct{} {
	return map[string]struct{}{
		darwin:  {},
		linux:   {},
		windows: {},
	}
}

func IsOS(k string) bool {
	_, ok := GOOSMap()[k]
	return ok
}

func GOARCHList() []string {
	return []string{amd64, arm64}
}
