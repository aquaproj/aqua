// Package profile provides unified profiling functionality for aqua CLI operations.
// It combines CPU profiling and execution tracing capabilities to enable
// comprehensive performance analysis of aqua commands.
package profile

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cpuprofile"
	"github.com/aquaproj/aqua/v2/pkg/cli/tracer"
	"github.com/urfave/cli/v3"
)

// Profiler manages both CPU profiling and execution tracing for aqua operations.
// It coordinates multiple profiling mechanisms to provide comprehensive
// performance analysis capabilities.
type Profiler struct {
	cpu    *cpuprofile.CPUProfiler
	tracer *tracer.Tracer
}

// Stop terminates all active profiling sessions and cleans up resources.
// It safely stops both CPU profiling and execution tracing,
// ensuring proper cleanup even if components are nil.
func (p *Profiler) Stop() {
	p.cpu.Stop()
	p.tracer.Stop()
}

// Start initializes profiling based on CLI command flags and returns a Profiler instance.
// It starts both execution tracing and CPU profiling if the respective flags are provided.
// If any profiling mechanism fails to start, it cleans up already started profilers.
func Start(cmd *cli.Command) (*Profiler, error) {
	t, err := tracer.Start(cmd.String("trace"))
	if err != nil {
		return nil, fmt.Errorf("start tracing: %w", err)
	}
	cpuProfiler, err := cpuprofile.Start(cmd.String("cpu-profile"))
	if err != nil {
		t.Stop()
		return nil, fmt.Errorf("start CPU Profile: %w", err)
	}
	return &Profiler{
		tracer: t,
		cpu:    cpuProfiler,
	}, nil
}
