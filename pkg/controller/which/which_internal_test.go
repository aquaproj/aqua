package which

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/pkgtype/http"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func Test_controller_findExecFileFromPkg(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title          string
		registries     map[string]*registry.Config
		exeName        string
		pkg            *aqua.Package
		expPackageInfo *registry.PackageInfo
		expFile        *registry.File
	}{
		{
			title:   "normal",
			exeName: "kubectl",
			pkg: &aqua.Package{
				Registry: "standard",
				Name:     "kubernetes/kubectl",
			},
			expPackageInfo: &registry.PackageInfo{
				Name: "kubernetes/kubectl",
				Type: "http",
				Files: []*registry.File{
					{
						Name: "kubectl",
					},
				},
			},
			expFile: &registry.File{
				Name: "kubectl",
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						&registry.PackageInfo{
							Name: "kubernetes/kubectl",
							Type: "http",
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
	ctrl := &controller{
		runtime: runtime.New(),
		installers: config.PackageTypes{
			http.PkgType: http.New(&config.Param{}, nil, nil),
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkg, file := ctrl.findExecFileFromPkg(d.registries, d.exeName, d.pkg, logE)
			if diff := cmp.Diff(d.expPackageInfo, pkg.PackageInfo); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(d.expFile, file); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
