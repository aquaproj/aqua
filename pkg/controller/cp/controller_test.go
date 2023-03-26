package cp_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/cp"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func boolP(b bool) *bool {
	return &b
}

func TestController_Copy(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		param        *config.Param
		pkgInstaller cp.PackageInstaller
		fs           afero.Fs
		rt           *runtime.Runtime
		whichCtrl    which.Controller
		installer    cp.Installer
		isErr        bool
	}{
		{
			name:      "no arg",
			param:     &config.Param{},
			fs:        afero.NewMemMapFs(),
			installer: &cp.MockInstaller{},
		},
		{
			name: "gh",
			param: &config.Param{
				MaxParallelism: 5,
				Args: []string{
					"gh",
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
			fs:        afero.NewMemMapFs(),
			installer: &cp.MockInstaller{},
			whichCtrl: &which.MockController{
				FindResult: &which.FindResult{
					ExePath: "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_macOS_amd64.tar.gz/gh_2.17.0_macOS_amd64/bin/gh",
					Package: &config.Package{
						Package: &aqua.Package{
							Name: "cli/cli",
						},
					},
					Config: &aqua.Config{
						Checksum: &aqua.Checksum{
							Enabled:         boolP(true),
							RequireChecksum: boolP(true),
						},
					},
					ConfigFilePath: "aqua.yaml",
				},
			},
			pkgInstaller: &cp.MockPackageInstaller{},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctrl := cp.New(d.param, d.pkgInstaller, d.fs, d.rt, d.whichCtrl, d.installer, &policy.MockConfigReader{}, &policy.MockConfigFinder{}, &policy.MockValidator{})
			if err := ctrl.Copy(ctx, logE, d.param); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error should be returned")
			}
		})
	}
}
