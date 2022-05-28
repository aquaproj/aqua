package installpackage_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/exec"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/link"
	"github.com/aquaproj/aqua/pkg/runtime"
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
		cfg        *config.Config
		registries map[string]*config.RegistryContent
		executor   exec.Executor
		binDir     string
		onlyLink   bool
		isTest     bool
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
			cfg: &config.Config{
				Registries: config.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*config.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v2.0.3",
					},
					{
						Name:     "suzuki-shunsuke/github-comment",
						Registry: "standard",
						Version:  "v4.1.0",
					},
				},
			},
			registries: map[string]*config.RegistryContent{
				"standard": {
					PackageInfos: config.PackageInfos{
						{
							Type:      "github_release",
							RepoOwner: "suzuki-shunsuke",
							RepoName:  "ci-info",
							Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
						},
						{
							Type:        "github_release",
							RepoOwner:   "suzuki-shunsuke",
							RepoName:    "github-comment",
							Asset:       stringP("github-comment_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
							SupportedIf: stringP("false"),
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
			name:     "only link",
			onlyLink: true,
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
			cfg: &config.Config{
				Registries: config.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*config.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v2.0.3",
					},
				},
			},
			registries: map[string]*config.RegistryContent{
				"standard": {
					PackageInfos: config.PackageInfos{
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
			cfg: &config.Config{
				Registries: config.Registries{
					"standard": {
						Name:      "standard",
						Type:      "github_content",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Ref:       "v2.15.0",
						Path:      "registry.yaml",
					},
				},
				Packages: []*config.Package{},
			},
			registries: map[string]*config.RegistryContent{
				"standard": {
					PackageInfos: config.PackageInfos{},
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
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			linker := link.NewMockLinker(afero.NewMemMapFs())
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewPackageDownloader(nil, d.rt, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, linker, d.executor)
			if err := ctrl.InstallPackages(ctx, d.cfg, d.registries, d.binDir, d.onlyLink, d.isTest, logE); err != nil {
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
		pkgInfo  *config.PackageInfo
		pkg      *config.Package
		executor exec.Executor
		isTest   bool
		isErr    bool
	}{
		{
			name: "file already exists",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			pkgInfo: &config.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     stringP("ci-info_{{trimV .Version}}_{{.OS}}_amd64.tar.gz"),
			},
			pkg: &config.Package{
				Name:     "suzuki-shunsuke/ci-info",
				Registry: "standard",
				Version:  "v2.0.3",
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
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewPackageDownloader(nil, d.rt, download.NewHTTPDownloader(http.DefaultClient))
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, nil, d.executor)
			if err := ctrl.InstallPackage(ctx, d.pkgInfo, d.pkg, d.isTest, logE); err != nil {
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
