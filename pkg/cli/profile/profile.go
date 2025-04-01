package profile

import (
	"fmt"

	"github.com/aquaproj/aqua/v2/pkg/cli/cpuprofile"
	"github.com/aquaproj/aqua/v2/pkg/cli/tracer"
	"github.com/urfave/cli/v3"
)

type Profiler struct {
	cpu    *cpuprofile.CPUProfiler
	tracer *tracer.Tracer
}

func (p *Profiler) Stop() {
	p.cpu.Stop()
	p.tracer.Stop()
}

func Start(c *cli.Context) (*Profiler, error) {
	t, err := tracer.Start(c.String("trace"))
	if err != nil {
		return nil, fmt.Errorf("start tracing: %w", err)
	}
	cpuProfiler, err := cpuprofile.Start(c.String("cpu-profile"))
	if err != nil {
		t.Stop()
		return nil, fmt.Errorf("start CPU Profile: %w", err)
	}
	return &Profiler{
		tracer: t,
		cpu:    cpuProfiler,
	}, nil
}
