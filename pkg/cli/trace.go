package cli

import (
	"fmt"
	"io"
	"os"
	"runtime/trace"
)

type Tracer struct {
	f io.Closer
}

func (t *Tracer) Stop() {
	if t == nil {
		return
	}
	trace.Stop()
	t.f.Close()
}

func startTrace(p string) (*Tracer, error) {
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
