package installpackage_test

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/spf13/afero"
)

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
			name: errFileAlreadyExists,
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			param: &config.Param{
				CWD:            pathWorkspace,
				ConfigFilePath: fileAquaYaml,
				RootDir:        pathRoot,
				MaxParallelism: 5,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: {
						Name:      regTypeStandard,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       versionV215,
						Path:      regFileRegistryYaml,
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     repoSuzukiShunsukeCiInfo,
						Registry: regTypeStandard,
						Version:  versionV203,
					},
				},
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoSuzukiShunsuke,
							RepoName:  repoNameCiInfo,
							Asset:     tmplCiInfoAsset,
						},
					},
				},
			},
			binDir: pathRootBin,
			files: map[string]string{
				pathCiInfoBinary: ``,
			},
			links: map[string]string{
				"../aqua-proxy": "/home/foo/.local/share/aquaproj-aqua/bin/ci-info",
			},
		},
		{
			name: "only link",
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			param: &config.Param{
				CWD:            pathWorkspace,
				ConfigFilePath: fileAquaYaml,
				RootDir:        pathRoot,
				MaxParallelism: 5,
				OnlyLink:       true,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: {
						Name:      regTypeStandard,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       versionV215,
						Path:      regFileRegistryYaml,
					},
				},
				Packages: []*aqua.Package{
					{
						Name:     repoSuzukiShunsukeCiInfo,
						Registry: regTypeStandard,
						Version:  versionV203,
					},
				},
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{
						{
							Type:      pkgTypeGitHubRelease,
							RepoOwner: repoSuzukiShunsuke,
							RepoName:  repoNameCiInfo,
							Asset:     tmplCiInfoAsset,
						},
					},
				},
			},
			binDir: pathRootBin,
			files: map[string]string{
				pathCiInfoBinary: ``,
			},
			links: map[string]string{
				"../aqua-proxy": "/home/foo/.local/share/aquaproj-aqua/bin/ci-info",
			},
		},
		{
			name: "no package",
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			param: &config.Param{
				CWD:            pathWorkspace,
				ConfigFilePath: fileAquaYaml,
				RootDir:        pathRoot,
				MaxParallelism: 5,
			},
			cfg: &aqua.Config{
				Registries: aqua.Registries{
					regTypeStandard: {
						Name:      regTypeStandard,
						Type:      pkgTypeGitHubContent,
						RepoOwner: regOwnerAquaproj,
						RepoName:  regNameAquaRegistry,
						Ref:       versionV215,
						Path:      regFileRegistryYaml,
					},
				},
				Packages: []*aqua.Package{},
			},
			registries: map[string]*registry.Config{
				regTypeStandard: {
					PackageInfos: registry.PackageInfos{},
				},
			},
			binDir: pathRootBin,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			linker := installpackage.NewMockLinker(afero.NewMemMapFs())
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			vacuumMock := vacuum.NewMock(d.param.RootDir, nil, nil)
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, linker, nil, &checksum.Calculator{}, unarchive.New(d.executor, fs), &cosign.MockVerifier{}, &slsa.MockVerifier{}, &minisign.MockVerifier{}, &ghattestation.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{}, vacuumMock)
			if err := ctrl.InstallPackages(ctx, logger, &installpackage.ParamInstallPackages{
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
			name: errFileAlreadyExists,
			rt: &runtime.Runtime{
				GOOS:   osLinux,
				GOARCH: archAmd64,
			},
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      pkgTypeGitHubRelease,
					RepoOwner: repoSuzukiShunsuke,
					RepoName:  repoNameCiInfo,
					Asset:     tmplCiInfoAsset,
				},
				Package: &aqua.Package{
					Name:     repoSuzukiShunsukeCiInfo,
					Registry: regTypeStandard,
					Version:  versionV203,
				},
				Registry: &aqua.Registry{
					Name:      regTypeStandard,
					Type:      pkgTypeGitHubContent,
					RepoOwner: regOwnerAquaproj,
					RepoName:  regNameAquaRegistry,
					Ref:       versionV215,
					Path:      regFileRegistryYaml,
				},
			},
			param: &config.Param{
				RootDir: pathRoot,
			},
			files: map[string]string{
				pathCiInfoBinary: ``,
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			vacuumMock := vacuum.NewMock(d.param.RootDir, nil, nil)
			ctrl := installpackage.New(d.param, downloader, d.rt, fs, nil, nil, &checksum.Calculator{}, unarchive.New(d.executor, fs), &cosign.MockVerifier{}, &slsa.MockVerifier{}, &minisign.MockVerifier{}, &ghattestation.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{}, vacuumMock)
			if err := ctrl.InstallPackage(ctx, logger, &installpackage.ParamInstallPackage{
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
