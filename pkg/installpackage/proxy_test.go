package installpackage_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubcontent"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubrelease"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func Test_installer_InstallProxy(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		files    map[string]string
		param    *config.Param
		rt       *runtime.Runtime
		executor exec.Executor
		links    map[string]string
		isErr    bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/aquaproj/aqua-proxy/v1.1.0/aqua-proxy_linux_amd64.tar.gz/aqua-proxy": "",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			linker := link.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			pkgTypes := config.PackageTypes{
				githubrelease.PkgType: githubrelease.New(d.param, fs, d.rt, nil),
				githubcontent.PkgType: githubcontent.New(d.param, fs, nil),
			}
			ctrl := installpackage.New(d.param, d.rt, fs, linker, d.executor, pkgTypes)
			if err := ctrl.InstallProxy(ctx, logE); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
