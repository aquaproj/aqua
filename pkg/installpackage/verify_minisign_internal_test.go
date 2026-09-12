package installpackage

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func TestMinisignVerifier_Enabled(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		// rt is the target platform, which AQUA_GOOS and AQUA_GOARCH can change.
		rt *runtime.Runtime
		// realRT is the host platform, where minisign is executed.
		realRT   *runtime.Runtime
		minisign *registry.Minisign
		exp      bool
	}{
		{
			name:     "disabled",
			rt:       &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			realRT:   &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			minisign: nil,
			exp:      false,
		},
		{
			name:     "supported",
			rt:       &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			realRT:   &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			minisign: &registry.Minisign{},
			exp:      true,
		},
		{
			// minisign provides no linux/arm64 asset, so the verification must be
			// skipped even if the target platform is linux/amd64.
			name:     "the host platform isn't supported",
			rt:       &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			realRT:   &runtime.Runtime{GOOS: "linux", GOARCH: "arm64"},
			minisign: &registry.Minisign{},
			exp:      false,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			v := &minisignVerifier{
				runtime:     d.rt,
				realRuntime: d.realRT,
				minisign:    d.minisign,
			}
			enabled, err := v.Enabled(logger)
			if err != nil {
				t.Fatal(err)
			}
			if enabled != d.exp {
				t.Fatalf("got %v, wanted %v", enabled, d.exp)
			}
		})
	}
}
