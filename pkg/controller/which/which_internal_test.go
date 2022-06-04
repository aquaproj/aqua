package which

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/pkgtype"
	"github.com/aquaproj/aqua/pkg/pkgtype/http"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

func Test_controller_findExecFileFromPkg(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title          string
		files          map[string]string
		registries     map[string]*config.RegistryContent
		exeName        string
		pkg            *config.Package
		expPackageInfo *config.PackageInfo
		expFile        *config.File
		param          *config.Param
		rt             *runtime.Runtime
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
				Type: "http",
				URL:  stringP("https://storage.googleapis.com/kubernetes-release/release/{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl"),
				Files: []*config.File{
					{
						Name: "kubectl",
					},
				},
			},
			expFile: &config.File{
				Name: "kubectl",
			},
			param: &config.Param{},
			registries: map[string]*config.RegistryContent{
				"standard": {
					PackageInfos: config.PackageInfos{
						&config.PackageInfo{
							Name: "kubernetes/kubectl",
							Type: "http",
							URL:  stringP("https://storage.googleapis.com/kubernetes-release/release/{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl"),
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
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			ctrl := &controller{
				runtime: runtime.New(),
				pkgTypes: pkgtype.Packages{
					http.PkgType: http.New(d.param, fs, d.rt),
				},
			}
			pkgInfo, file, _ := ctrl.findExecFileFromPkg(d.registries, d.exeName, d.pkg, logE)
			if diff := cmp.Diff(d.expPackageInfo, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(d.expFile, file); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
