package which

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func Test_controller_findExecFileFromPkg(t *testing.T) {
	t.Parallel()
	data := []struct {
		title          string
		registries     map[string]*config.RegistryContent
		exeName        string
		pkg            *config.Package
		expPackageInfo *config.PackageInfo
		expFile        *config.File
	}{
		{
			title:   "normal",
			exeName: "kubectl",
			pkg: &config.Package{
				Registry: "standard",
				Name:     "kubernetes/kubectl",
			},
			expPackageInfo: &config.PackageInfo{
				Name: "kubernetes/kubectl",
				Files: []*config.File{
					{
						Name: "kubectl",
					},
				},
			},
			expFile: &config.File{
				Name: "kubectl",
			},
			registries: map[string]*config.RegistryContent{
				"standard": {
					PackageInfos: config.PackageInfos{
						&config.PackageInfo{
							Name: "kubernetes/kubectl",
							Files: []*config.File{
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
