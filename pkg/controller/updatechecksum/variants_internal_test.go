package updatechecksum

import (
	"slices"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func sortedKeys(rts []*runtime.Runtime) []string {
	keys := make([]string, 0, len(rts))
	for _, rt := range rts {
		keys = append(keys, runtimeKey(rt))
	}
	slices.Sort(keys)
	return keys
}

func TestExpandRuntimesByVariants(t *testing.T) { //nolint:funlen
	t.Parallel()
	linuxAmd64 := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	linuxArm64 := &runtime.Runtime{GOOS: "linux", GOARCH: "arm64"}
	darwinArm64 := &runtime.Runtime{GOOS: "darwin", GOARCH: "arm64"}

	tests := []struct {
		name     string
		pkgInfo  *registry.PackageInfo
		rts      []*runtime.Runtime
		wantKeys []string
	}{
		{
			name:     "no overrides is a no-op",
			pkgInfo:  &registry.PackageInfo{},
			rts:      []*runtime.Runtime{linuxAmd64, darwinArm64},
			wantKeys: []string{"darwin/arm64/", "linux/amd64/"},
		},
		{
			name: "overrides without variants is a no-op",
			pkgInfo: &registry.PackageInfo{
				Overrides: registry.Overrides{
					{GOOS: "linux", GOArch: "amd64"},
				},
			},
			rts:      []*runtime.Runtime{linuxAmd64, darwinArm64},
			wantKeys: []string{"darwin/arm64/", "linux/amd64/"},
		},
		{
			name: "musl and glibc both expand for linux/amd64",
			pkgInfo: &registry.PackageInfo{
				Overrides: registry.Overrides{
					{
						GOOS:     "linux",
						GOArch:   "amd64",
						Variants: registry.Variants{{Key: "libc", Value: "musl"}},
					},
					{
						GOOS:     "linux",
						GOArch:   "amd64",
						Variants: registry.Variants{{Key: "libc", Value: "glibc"}},
					},
				},
			},
			rts:      []*runtime.Runtime{linuxAmd64, darwinArm64},
			wantKeys: []string{"darwin/arm64/", "linux/amd64/glibc", "linux/amd64/musl"},
		},
		{
			name: "musl-only override drops empty-LibC runtime for that platform",
			pkgInfo: &registry.PackageInfo{
				Overrides: registry.Overrides{
					{
						GOOS:     "linux",
						GOArch:   "amd64",
						Variants: registry.Variants{{Key: "libc", Value: "musl"}},
					},
				},
			},
			rts:      []*runtime.Runtime{linuxAmd64, linuxArm64},
			wantKeys: []string{"linux/amd64/musl", "linux/arm64/"},
		},
		{
			name: "fallback override coexists with libc-constrained override",
			pkgInfo: &registry.PackageInfo{
				Overrides: registry.Overrides{
					{
						GOOS:     "linux",
						GOArch:   "amd64",
						Variants: registry.Variants{{Key: "libc", Value: "musl"}},
					},
					{GOOS: "linux", GOArch: "amd64"},
				},
			},
			rts:      []*runtime.Runtime{linuxAmd64},
			wantKeys: []string{"linux/amd64/", "linux/amd64/musl"},
		},
		{
			name: "override with unsupported variant key is ignored",
			pkgInfo: &registry.PackageInfo{
				Overrides: registry.Overrides{
					{
						GOOS:     "linux",
						GOArch:   "amd64",
						Variants: registry.Variants{{Key: "foo", Value: "bar"}},
					},
				},
			},
			rts:      []*runtime.Runtime{linuxAmd64},
			wantKeys: []string{"linux/amd64/"},
		},
		{
			name: "variant override scoped by GOOS only spans all matching archs",
			pkgInfo: &registry.PackageInfo{
				Overrides: registry.Overrides{
					{
						GOOS:     "linux",
						Variants: registry.Variants{{Key: "libc", Value: "musl"}},
					},
				},
			},
			rts:      []*runtime.Runtime{linuxAmd64, linuxArm64, darwinArm64},
			wantKeys: []string{"darwin/arm64/", "linux/amd64/musl", "linux/arm64/musl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := sortedKeys(expandRuntimesByVariants(tt.pkgInfo, tt.rts))
			if !slices.Equal(got, tt.wantKeys) {
				t.Errorf("expandRuntimesByVariants() keys = %v, want %v", got, tt.wantKeys)
			}
		})
	}
}
