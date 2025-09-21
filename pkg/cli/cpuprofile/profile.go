// Package cpuprofile provides CPU profiling functionality for aqua CLI operations.
// It wraps Go's runtime pprof capabilities to enable performance analysis
// and optimization of aqua command execution.
package cpuprofile

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"
)

// CPUProfiler manages CPU profiling for aqua operations.
// It holds a file handle for writing profile output and provides
// methods to start and stop profiling sessions.
type CPUProfiler struct {
	f io.Closer
}

// Stop terminates the CPU profiling session and closes the output file.
// It safely handles nil receivers and ensures proper cleanup
// of profiling resources.
func (cp *CPUProfiler) Stop() {
	if cp == nil {
		return
	}
	pprof.StopCPUProfile()
	cp.f.Close()
}

// Start begins CPU profiling and writes profile data to the specified file path.
// If the path is empty, no profiling is started and nil is returned.
// Returns a CPUProfiler instance that should be used to stop profiling when complete.
func Start(p string) (*CPUProfiler, error) {
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
