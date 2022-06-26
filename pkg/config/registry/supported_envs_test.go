package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
)

func TestPackageInfo_CheckSupported(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		pkgInfo *registry.PackageInfo
		rt      *runtime.Runtime
		exp     bool
		isErr   bool
	}{
		{
			name:    "empty",
			pkgInfo: &registry.PackageInfo{},
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: true,
		},
		{
			name: "supported_envs os true",
			pkgInfo: &registry.PackageInfo{
				SupportedEnvs: registry.SupportedEnvs{
					"linux",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: true,
		},
		{
			name: "supported_envs arch true",
			pkgInfo: &registry.PackageInfo{
				SupportedEnvs: registry.SupportedEnvs{
					"darwin",
					"amd64",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: true,
		},
		{
			name: "supported_envs os/arch true",
			pkgInfo: &registry.PackageInfo{
				SupportedEnvs: registry.SupportedEnvs{
					"darwin",
					"arm64",
					"linux/amd64",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: true,
		},
		{
			name: "supported_envs os false",
			pkgInfo: &registry.PackageInfo{
				SupportedEnvs: registry.SupportedEnvs{
					"darwin",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: false,
		},
		{
			name: "supported_envs all",
			pkgInfo: &registry.PackageInfo{
				SupportedEnvs: registry.SupportedEnvs{
					"all",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			exp: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			b, err := d.pkgInfo.CheckSupported(d.rt, d.rt.GOOS+"/"+d.rt.GOARCH)
			if d.isErr {
				if err == nil {
					t.Fatal("error must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if d.exp != b {
				t.Fatalf("wanted %v, got %v", d.exp, b)
			}
		})
	}
}
