package cp_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/controller/cp"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type mockInstaller struct {
	err error
}

func (inst *mockInstaller) Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	return inst.err
}

type mockWhichController struct {
	findResult *domain.FindResult
	err        error
}

func (ctrl *mockWhichController) Which(ctx context.Context, param *config.Param, exeName string, logE *logrus.Entry) (*domain.FindResult, error) {
	return ctrl.findResult, ctrl.err
}

type mockPackageInstaller struct{}

func (inst *mockPackageInstaller) InstallPackage(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackage) error {
	return nil
}

func (inst *mockPackageInstaller) InstallPackages(ctx context.Context, logE *logrus.Entry, param *domain.ParamInstallPackages) error {
	return nil
}

func (inst *mockPackageInstaller) SetCopyDir(copyDir string) {
}

func (inst *mockPackageInstaller) Copy(dest, src string) error {
	return nil
}

func (inst *mockPackageInstaller) WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error {
	return nil
}

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
		whichCtrl    domain.WhichController
		installer    cp.Installer
		isErr        bool
	}{
		{
			name:      "no arg",
			param:     &config.Param{},
			fs:        afero.NewMemMapFs(),
			installer: &mockInstaller{},
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
			installer: &mockInstaller{},
			whichCtrl: &mockWhichController{
				findResult: &domain.FindResult{
					ExePath: "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_macOS_amd64.tar.gz/gh_2.17.0_macOS_amd64/bin/gh",
					Package: &config.Package{
						Package: &aqua.Package{
							Name: "cli/cli",
						},
					},
					Config: &aqua.Config{
						Checksum: &aqua.Checksum{
							Enabled:         boolP(true),
							RequireChecksum: true,
						},
					},
					ConfigFilePath: "aqua.yaml",
				},
			},
			pkgInstaller: &mockPackageInstaller{},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctrl := cp.New(d.param, d.pkgInstaller, d.fs, d.rt, d.whichCtrl, d.installer)
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
