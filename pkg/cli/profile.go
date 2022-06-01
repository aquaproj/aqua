package cli

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"
)

type CPUProfiler struct {
	f io.Closer
}

func (cp *CPUProfiler) Stop() {
	if cp == nil {
		return
	}
	pprof.StopCPUProfile()
	cp.f.Close()
}

func startCPUProfile(p string) (*CPUProfiler, error) {
	if p == "" {
		return nil, nil //nolint:nilnil
	}
	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("create a cpu profile output file: %w", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("start a trace: %w", err)
	}
	return &CPUProfiler{
		f: f,
	}, nil
}
