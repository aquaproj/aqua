package which

import (
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func strP(s string) *string {
	return &s
}

func TestController_findExecFileFromPkg(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title      string
		registries map[string]*registry.Config
		exeName    string
		pkg        *aqua.Package
		expWhich   *domain.FindResult
	}{
		{
			title:   "normal",
			exeName: "kubectl",
			pkg: &aqua.Package{
				Registry: "standard",
				Name:     "kubernetes/kubectl",
				Version:  "v1.21.0",
			},
			expWhich: &domain.FindResult{
				Package: &config.Package{
					Package: &aqua.Package{
						Registry: "standard",
						Name:     "kubernetes/kubectl",
						Version:  "v1.21.0",
					},
					PackageInfo: &registry.PackageInfo{
						Type: "http",
						Name: "kubernetes/kubectl",
						URL:  strP("https://storage.googleapis.com/kubernetes-release/release/{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl"),
						Files: []*registry.File{
							{
								Name: "kubectl",
							},
						},
					},
				},
				File: &registry.File{
					Name: "kubectl",
				},
				ExePath: filepath.Join("/home", "foo", ".local", "share", "aquaproj-aqua", "pkgs", "http", "storage.googleapis.com/kubernetes-release/release/v1.21.0/bin/linux/amd64/kubectl/kubectl"),
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						&registry.PackageInfo{
							Type: "http",
							Name: "kubernetes/kubectl",
							URL:  strP("https://storage.googleapis.com/kubernetes-release/release/{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl"),
							Files: []*registry.File{
								{
									Name: "kubectl",
								},
							},
						},
					},
				},
			},
		},
	}
	ctrl := &Controller{
		runtime: &runtime.Runtime{
			GOOS:   "linux",
			GOARCH: "amd64",
		},
		rootDir: filepath.Join("/home", "foo", ".local", "share", "aquaproj-aqua"),
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			which := ctrl.findExecFileFromPkg(d.registries, d.exeName, d.pkg, logE)
			if diff := cmp.Diff(d.expWhich, which); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
