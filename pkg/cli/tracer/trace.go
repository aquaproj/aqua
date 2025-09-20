// Package tracer provides execution tracing functionality for aqua CLI operations.
// It wraps Go's runtime tracing capabilities to help with performance analysis
// and debugging of aqua command execution.
package tracer

import (
	"fmt"
	"io"
	"os"
	"runtime/trace"
)

// Tracer manages execution tracing for aqua operations.
// It holds a file handle for writing trace output and provides
// methods to start and stop tracing sessions.
type Tracer struct {
	f io.Closer
}

// Stop terminates the tracing session and closes the output file.
// It safely handles nil receivers and ensures proper cleanup
// of tracing resources.
func (t *Tracer) Stop() {
	if t == nil {
		return
	}
	trace.Stop()
	t.f.Close()
}

// Start begins execution tracing and writes trace data to the specified file path.
// If the path is empty, no tracing is started and nil is returned.
// Returns a Tracer instance that should be used to stop tracing when complete.
func Start(p string) (*Tracer, error) {
	if p == "" {
		return nil, nil //nolint:nilnil
	}
	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("create a trace output file: %w", err)
	}
	if err := trace.Start(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("start a trace: %w", err)
	}
	return &Tracer{
		f: f,
	}, nil
}
