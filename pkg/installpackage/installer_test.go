package installpackage_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

func Test_installer_InstallPackages(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		files      map[string]string
		links      map[string]string
		param      *config.Param
		rt         *runtime.Runtime
		cfg        *aqua.Config
		registries map[string]*registry.Config
		executor   installpackage.Executor
		binDir     string
		isErr      bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v2.0.3",
					},
				},
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "ci-info",
							Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
						},
					},
				},
			},
			binDir: "/home/foo/.local/share/aquaproj-aqua/bin",
			files: map[string]string{
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info": ``,
			},
			links: map[string]string{
				"aqua-proxy": "/home/foo/.local/share/aquaproj-aqua/bin/ci-info",
			},
		},
		{
			name: "only link",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
				OnlyLink:       true,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v2.0.3",
					},
				},
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "ci-info",
							Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
						},
					},
				},
			},
			binDir: "/home/foo/.local/share/aquaproj-aqua/bin",
			files: map[string]string{
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info": ``,
			},
			links: map[string]string{
				"aqua-proxy": "/home/foo/.local/share/aquaproj-aqua/bin/ci-info",
			},
		},
		{
			name: "no package",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*aqua.Package{},
			},
			registries: map[string]*registry.Config{
				"standard": {
					PackageInfos: registry.PackageInfos{},
				},
			},
			binDir: "/home/foo/.local/share/aquaproj-aqua/bin",
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			linker := domain.NewMockLinker(afero.NewMemMapFs())
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, linker, d.executor, nil, &checksum.Calculator{}, unarchive.New(d.executor), &policy.MockChecker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
			if err := ctrl.InstallPackages(ctx, logE, &installpackage.ParamInstallPackages{
				Config:         d.cfg,
				Registries:     d.registries,
				ConfigFilePath: d.param.ConfigFilePath,
			}); err != nil {
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

func Test_installer_InstallPackage(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		files    map[string]string
		param    *config.Param
		rt       *runtime.Runtime
		pkg      *config.Package
		executor installpackage.Executor
		isTest   bool
		isErr    bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "suzuki-shunsuke",
					RepoName:  "ci-info",
					Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
				},
				Package: &aqua.Package{
					Name:     "suzuki-shunsuke/ci-info",
					Registry: "standard",
					Version:  "v2.0.3",
				},
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			files: map[string]string{
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/suzuki-shunsuke/ci-info/v2.0.3/ci-info_2.0.3_linux_amd64.tar.gz/ci-info": ``,
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, nil, d.executor, nil, &checksum.Calculator{}, unarchive.New(d.executor), &policy.MockChecker{}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
			if err := ctrl.InstallPackage(ctx, logE, &installpackage.ParamInstallPackage{
				Pkg: d.pkg,
			}); err != nil {
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
