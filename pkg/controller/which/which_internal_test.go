package which

import (
	"testing"

	"github.com/clivm/clivm/pkg/config/aqua"
	"github.com/clivm/clivm/pkg/config/registry"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func Test_controller_findExecFileFromPkg(t *testing.T) {
	t.Parallel()
	data := []struct {
		title          string
		registries     map[string]*registry.Config
		exeName        string
		pkg            *clivm.Package
		expPackageInfo *registry.PackageInfo
		expFile        *registry.File
	}{
		{
			title:   "normal",
			exeName: "kubectl",
			pkg: &clivm.Package{
				Registry: "standard",
				Name:     "kubernetes/kubectl",
			},
			expPackageInfo: &registry.PackageInfo{
				Name: "kubernetes/kubectl",
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
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, file := ctrl.findExecFileFromPkg(d.registries, d.exeName, d.pkg, logE)
			if diff := cmp.Diff(d.expPackageInfo, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(d.expFile, file); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
