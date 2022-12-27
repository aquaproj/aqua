package runtime

import (
	"os"
	"runtime"
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
	return []string{"darwin", "linux", "windows"}
}

func GOARCHList() []string {
	return []string{"amd64", "arm64"}
}
