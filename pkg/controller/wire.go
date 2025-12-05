//go:build wireinject
// +build wireinject

package controller

import (
	"context"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/controller/allowpolicy"
	"github.com/aquaproj/aqua/v2/pkg/controller/cp"
	"github.com/aquaproj/aqua/v2/pkg/controller/denypolicy"
	cexec "github.com/aquaproj/aqua/v2/pkg/controller/exec"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate"
	genrgst "github.com/aquaproj/aqua/v2/pkg/controller/generate-registry"
	"github.com/aquaproj/aqua/v2/pkg/controller/generate/output"
	"github.com/aquaproj/aqua/v2/pkg/controller/info"
	"github.com/aquaproj/aqua/v2/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/v2/pkg/controller/initpolicy"
	"github.com/aquaproj/aqua/v2/pkg/controller/install"
	"github.com/aquaproj/aqua/v2/pkg/controller/list"
	"github.com/aquaproj/aqua/v2/pkg/controller/remove"
	"github.com/aquaproj/aqua/v2/pkg/controller/update"
	"github.com/aquaproj/aqua/v2/pkg/controller/updateaqua"
	"github.com/aquaproj/aqua/v2/pkg/controller/updatechecksum"
	cvacuum "github.com/aquaproj/aqua/v2/pkg/controller/vacuum"
	"github.com/aquaproj/aqua/v2/pkg/controller/vacuum/initialize"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/aquaproj/aqua/v2/pkg/github"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/goproxy"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func InitializeListCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *list.Controller {
	wire.Build(
		list.NewController,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(list.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(list.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(list.ConfigReader), new(*reader.ConfigReader)),
		),
		afero.NewOsFs,
		download.NewHTTPDownloader,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return &list.Controller{}
}

func InitializeGenerateRegistryCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, stdout io.Writer) *genrgst.Controller {
	wire.Build(
		genrgst.NewController,
		wire.NewSet(
			github.New,
			wire.Bind(new(genrgst.RepositoriesService), new(*github.RepositoriesService)),
		),
		afero.NewOsFs,
		wire.NewSet(
			output.New,
			wire.Bind(new(genrgst.TestdataOutputter), new(*output.Outputter)),
		),
		wire.NewSet(
			cargo.NewClient,
			wire.Bind(new(genrgst.CargoClient), new(*cargo.Client)),
		),
	)
	return &genrgst.Controller{}
}

func InitializeInitCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param) *initcmd.Controller {
	wire.Build(
		initcmd.New,
		wire.NewSet(
			github.New,
			wire.Bind(new(initcmd.RepositoriesService), new(*github.RepositoriesService)),
		),
		afero.NewOsFs,
	)
	return &initcmd.Controller{}
}

func InitializeInitPolicyCommandController(ctx context.Context) *initpolicy.Controller {
	wire.Build(
		initpolicy.New,
		afero.NewOsFs,
	)
	return &initpolicy.Controller{}
}

func InitializeGenerateCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *generate.Controller {
	wire.Build(
		generate.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(generate.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(generate.RepositoriesService), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(versiongetter.GitHubTagClient), new(*github.RepositoriesService)),
			wire.Bind(new(versiongetter.GitHubReleaseClient), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(generate.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(generate.ConfigReader), new(*reader.ConfigReader)),
		),
		afero.NewOsFs,
		wire.NewSet(
			fuzzyfinder.New,
			wire.Bind(new(generate.FuzzyFinder), new(*fuzzyfinder.Finder)),
			wire.Bind(new(versiongetter.FuzzyFinder), new(*fuzzyfinder.Finder)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(installpackage.Executor), new(*osexec.Executor)),
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			cargo.NewClient,
			wire.Bind(new(versiongetter.CargoClient), new(*cargo.Client)),
		),
		wire.NewSet(
			versiongetter.NewFuzzy,
			wire.Bind(new(generate.FuzzyGetter), new(*versiongetter.FuzzyGetter)),
		),
		wire.NewSet(
			versiongetter.NewGeneralVersionGetter,
			wire.Bind(new(versiongetter.VersionGetter), new(*versiongetter.GeneralVersionGetter)),
		),
		versiongetter.NewCargo,
		versiongetter.NewGitHubRelease,
		versiongetter.NewGitHubTag,
		versiongetter.NewGoGetter,
		wire.NewSet(
			goproxy.New,
			wire.Bind(new(versiongetter.GoProxyClient), new(*goproxy.Client)),
		),
	)
	return &generate.Controller{}
}

func InitializeInstallCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) (*install.Controller, error) {
	wire.Build(
		install.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(install.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(install.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(install.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			installpackage.New,
			wire.Bind(new(install.Installer), new(*installpackage.Installer)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			afero.NewOsFs,
			wire.Bind(new(installpackage.Cleaner), new(afero.Fs)),
		),
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			osexec.New,
			wire.Bind(new(installpackage.Executor), new(*osexec.Executor)),
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(minisign.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(ghattestation.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewChecksumDownloader,
			wire.Bind(new(download.ChecksumDownloader), new(*download.ChecksumDownloaderImpl)),
		),
		wire.NewSet(
			checksum.NewCalculator,
			wire.Bind(new(installpackage.ChecksumCalculator), new(*checksum.Calculator)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(installpackage.Unarchiver), new(*unarchive.Unarchiver)),
		),
		wire.NewSet(
			policy.NewConfigReader,
			wire.Bind(new(policy.ConfigReader), new(*policy.ConfigReaderImpl)),
		),
		wire.NewSet(
			policy.NewConfigFinder,
			wire.Bind(new(policy.ConfigFinder), new(*policy.ConfigFinderImpl)),
		),
		wire.NewSet(
			policy.NewValidator,
			wire.Bind(new(policy.Validator), new(*policy.ValidatorImpl)),
		),
		wire.NewSet(
			policy.NewReader,
			wire.Bind(new(install.PolicyReader), new(*policy.Reader)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.New,
			wire.Bind(new(installpackage.MinisignVerifier), new(*minisign.Verifier)),
		),
		wire.NewSet(
			ghattestation.New,
			wire.Bind(new(installpackage.GitHubArtifactAttestationsVerifier), new(*ghattestation.Verifier)),
		),
		wire.NewSet(
			ghattestation.NewExecutor,
			wire.Bind(new(ghattestation.Executor), new(*ghattestation.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.NewExecutor,
			wire.Bind(new(minisign.Executor), new(*minisign.ExecutorImpl)),
		),
		wire.NewSet(
			installpackage.NewGoInstallInstallerImpl,
			wire.Bind(new(installpackage.GoInstallInstaller), new(*installpackage.GoInstallInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewGoBuildInstallerImpl,
			wire.Bind(new(installpackage.GoBuildInstaller), new(*installpackage.GoBuildInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewCargoPackageInstallerImpl,
			wire.Bind(new(installpackage.CargoPackageInstaller), new(*installpackage.CargoPackageInstallerImpl)),
		),
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(installpackage.Vacuum), new(*vacuum.Client)),
		),
	)
	return &install.Controller{}, nil
}

func InitializeWhichCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *which.Controller {
	wire.Build(
		which.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(which.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(which.ConfigReader), new(*reader.ConfigReader)),
		),
		osenv.New,
		afero.NewOsFs,
		download.NewHTTPDownloader,
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
			wire.Bind(new(which.Linker), new(*link.Linker)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return nil
}

func InitializeExecCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) (*cexec.Controller, error) {
	wire.Build(
		cexec.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			installpackage.New,
			wire.Bind(new(cexec.Installer), new(*installpackage.Installer)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(which.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(which.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(cexec.WhichController), new(*which.Controller)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(installpackage.Executor), new(*osexec.Executor)),
			wire.Bind(new(cexec.Executor), new(*osexec.Executor)),
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(minisign.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(ghattestation.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewChecksumDownloader,
			wire.Bind(new(download.ChecksumDownloader), new(*download.ChecksumDownloaderImpl)),
		),
		osenv.New,
		wire.NewSet(
			afero.NewOsFs,
			wire.Bind(new(installpackage.Cleaner), new(afero.Fs)),
		),
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
			wire.Bind(new(which.Linker), new(*link.Linker)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			checksum.NewCalculator,
			wire.Bind(new(installpackage.ChecksumCalculator), new(*checksum.Calculator)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(installpackage.Unarchiver), new(*unarchive.Unarchiver)),
		),
		wire.NewSet(
			policy.NewConfigReader,
			wire.Bind(new(policy.ConfigReader), new(*policy.ConfigReaderImpl)),
		),
		wire.NewSet(
			policy.NewConfigFinder,
			wire.Bind(new(policy.ConfigFinder), new(*policy.ConfigFinderImpl)),
		),
		wire.NewSet(
			policy.NewValidator,
			wire.Bind(new(policy.Validator), new(*policy.ValidatorImpl)),
		),
		wire.NewSet(
			policy.NewReader,
			wire.Bind(new(cexec.PolicyReader), new(*policy.Reader)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.New,
			wire.Bind(new(installpackage.MinisignVerifier), new(*minisign.Verifier)),
		),
		wire.NewSet(
			minisign.NewExecutor,
			wire.Bind(new(minisign.Executor), new(*minisign.ExecutorImpl)),
		),
		wire.NewSet(
			ghattestation.New,
			wire.Bind(new(installpackage.GitHubArtifactAttestationsVerifier), new(*ghattestation.Verifier)),
		),
		wire.NewSet(
			ghattestation.NewExecutor,
			wire.Bind(new(ghattestation.Executor), new(*ghattestation.ExecutorImpl)),
		),
		wire.NewSet(
			installpackage.NewGoInstallInstallerImpl,
			wire.Bind(new(installpackage.GoInstallInstaller), new(*installpackage.GoInstallInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewGoBuildInstallerImpl,
			wire.Bind(new(installpackage.GoBuildInstaller), new(*installpackage.GoBuildInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewCargoPackageInstallerImpl,
			wire.Bind(new(installpackage.CargoPackageInstaller), new(*installpackage.CargoPackageInstallerImpl)),
		),
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(installpackage.Vacuum), new(*vacuum.Client)),
			wire.Bind(new(cexec.Vacuum), new(*vacuum.Client)),
		),
	)
	return &cexec.Controller{}, nil
}

func InitializeUpdateAquaCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) (*updateaqua.Controller, error) {
	wire.Build(
		updateaqua.New,
		wire.NewSet(
			afero.NewOsFs,
			wire.Bind(new(installpackage.Cleaner), new(afero.Fs)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(updateaqua.RepositoriesService), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			installpackage.New,
			wire.Bind(new(updateaqua.AquaInstaller), new(*installpackage.Installer)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(installpackage.Executor), new(*osexec.Executor)),
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(minisign.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(ghattestation.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*osexec.Executor)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(installpackage.Unarchiver), new(*unarchive.Unarchiver)),
		),
		wire.NewSet(
			checksum.NewCalculator,
			wire.Bind(new(installpackage.ChecksumCalculator), new(*checksum.Calculator)),
		),
		wire.NewSet(
			download.NewChecksumDownloader,
			wire.Bind(new(download.ChecksumDownloader), new(*download.ChecksumDownloaderImpl)),
		),
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.New,
			wire.Bind(new(installpackage.MinisignVerifier), new(*minisign.Verifier)),
		),
		wire.NewSet(
			ghattestation.New,
			wire.Bind(new(installpackage.GitHubArtifactAttestationsVerifier), new(*ghattestation.Verifier)),
		),
		wire.NewSet(
			ghattestation.NewExecutor,
			wire.Bind(new(ghattestation.Executor), new(*ghattestation.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.NewExecutor,
			wire.Bind(new(minisign.Executor), new(*minisign.ExecutorImpl)),
		),
		wire.NewSet(
			installpackage.NewGoInstallInstallerImpl,
			wire.Bind(new(installpackage.GoInstallInstaller), new(*installpackage.GoInstallInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewGoBuildInstallerImpl,
			wire.Bind(new(installpackage.GoBuildInstaller), new(*installpackage.GoBuildInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewCargoPackageInstallerImpl,
			wire.Bind(new(installpackage.CargoPackageInstaller), new(*installpackage.CargoPackageInstallerImpl)),
		),
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(installpackage.Vacuum), new(*vacuum.Client)),
		),
	)
	return &updateaqua.Controller{}, nil
}

func InitializeCopyCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) (*cp.Controller, error) {
	wire.Build(
		cp.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
			wire.Bind(new(install.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			install.New,
			wire.Bind(new(cp.Installer), new(*install.Controller)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			installpackage.New,
			wire.Bind(new(install.Installer), new(*installpackage.Installer)),
			wire.Bind(new(cp.PackageInstaller), new(*installpackage.Installer)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(install.RegistryInstaller), new(*registry.Installer)),
			wire.Bind(new(which.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(which.ConfigReader), new(*reader.ConfigReader)),
			wire.Bind(new(install.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(cp.WhichController), new(*which.Controller)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(installpackage.Executor), new(*osexec.Executor)),
			wire.Bind(new(cexec.Executor), new(*osexec.Executor)),
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(minisign.CommandExecutor), new(*osexec.Executor)),
			wire.Bind(new(ghattestation.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewChecksumDownloader,
			wire.Bind(new(download.ChecksumDownloader), new(*download.ChecksumDownloaderImpl)),
		),
		osenv.New,
		wire.NewSet(
			afero.NewOsFs,
			wire.Bind(new(installpackage.Cleaner), new(afero.Fs)),
		),
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
			wire.Bind(new(which.Linker), new(*link.Linker)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			checksum.NewCalculator,
			wire.Bind(new(installpackage.ChecksumCalculator), new(*checksum.Calculator)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(installpackage.Unarchiver), new(*unarchive.Unarchiver)),
		),
		wire.NewSet(
			policy.NewConfigReader,
			wire.Bind(new(policy.ConfigReader), new(*policy.ConfigReaderImpl)),
		),
		wire.NewSet(
			policy.NewConfigFinder,
			wire.Bind(new(policy.ConfigFinder), new(*policy.ConfigFinderImpl)),
		),
		wire.NewSet(
			policy.NewValidator,
			wire.Bind(new(policy.Validator), new(*policy.ValidatorImpl)),
		),
		wire.NewSet(
			policy.NewReader,
			wire.Bind(new(cp.PolicyReader), new(*policy.Reader)),
			wire.Bind(new(install.PolicyReader), new(*policy.Reader)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.New,
			wire.Bind(new(installpackage.MinisignVerifier), new(*minisign.Verifier)),
		),
		wire.NewSet(
			ghattestation.New,
			wire.Bind(new(installpackage.GitHubArtifactAttestationsVerifier), new(*ghattestation.Verifier)),
		),
		wire.NewSet(
			ghattestation.NewExecutor,
			wire.Bind(new(ghattestation.Executor), new(*ghattestation.ExecutorImpl)),
		),
		wire.NewSet(
			minisign.NewExecutor,
			wire.Bind(new(minisign.Executor), new(*minisign.ExecutorImpl)),
		),
		wire.NewSet(
			installpackage.NewGoInstallInstallerImpl,
			wire.Bind(new(installpackage.GoInstallInstaller), new(*installpackage.GoInstallInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewGoBuildInstallerImpl,
			wire.Bind(new(installpackage.GoBuildInstaller), new(*installpackage.GoBuildInstallerImpl)),
		),
		wire.NewSet(
			installpackage.NewCargoPackageInstallerImpl,
			wire.Bind(new(installpackage.CargoPackageInstaller), new(*installpackage.CargoPackageInstallerImpl)),
		),
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(installpackage.Vacuum), new(*vacuum.Client)),
		),
	)
	return &cp.Controller{}, nil
}

func InitializeUpdateChecksumCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *updatechecksum.Controller {
	wire.Build(
		updatechecksum.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(updatechecksum.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(updatechecksum.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			download.NewChecksumDownloader,
			wire.Bind(new(download.ChecksumDownloader), new(*download.ChecksumDownloaderImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(updatechecksum.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
			wire.Bind(new(updatechecksum.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		afero.NewOsFs,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return &updatechecksum.Controller{}
}

func InitializeUpdateCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *update.Controller {
	wire.Build(
		update.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(update.ConfigFinder), new(*finder.ConfigFinder)),
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(update.ConfigReader), new(*reader.ConfigReader)),
			wire.Bind(new(which.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(update.RegistryInstaller), new(*registry.Installer)),
			wire.Bind(new(which.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(update.RepositoriesService), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
			wire.Bind(new(versiongetter.GitHubTagClient), new(*github.RepositoriesService)),
			wire.Bind(new(versiongetter.GitHubReleaseClient), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		afero.NewOsFs,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			versiongetter.NewFuzzy,
			wire.Bind(new(update.FuzzyGetter), new(*versiongetter.FuzzyGetter)),
		),
		wire.NewSet(
			fuzzyfinder.New,
			wire.Bind(new(update.FuzzyFinder), new(*fuzzyfinder.Finder)),
			wire.Bind(new(versiongetter.FuzzyFinder), new(*fuzzyfinder.Finder)),
		),
		wire.NewSet(
			versiongetter.NewGeneralVersionGetter,
			wire.Bind(new(versiongetter.VersionGetter), new(*versiongetter.GeneralVersionGetter)),
		),
		versiongetter.NewCargo,
		versiongetter.NewGitHubRelease,
		versiongetter.NewGitHubTag,
		versiongetter.NewGoGetter,
		wire.NewSet(
			cargo.NewClient,
			wire.Bind(new(versiongetter.CargoClient), new(*cargo.Client)),
		),
		wire.NewSet(
			goproxy.New,
			wire.Bind(new(versiongetter.GoProxyClient), new(*goproxy.Client)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(update.WhichController), new(*which.Controller)),
		),
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
			wire.Bind(new(which.Linker), new(*link.Linker)),
		),
		osenv.New,
	)
	return &update.Controller{}
}

func InitializeAllowPolicyCommandController(ctx context.Context, param *config.Param) *allowpolicy.Controller {
	wire.Build(
		allowpolicy.New,
		afero.NewOsFs,
		wire.NewSet(
			policy.NewConfigFinder,
			wire.Bind(new(policy.ConfigFinder), new(*policy.ConfigFinderImpl)),
		),
		wire.NewSet(
			policy.NewValidator,
			wire.Bind(new(policy.Validator), new(*policy.ValidatorImpl)),
		),
	)
	return &allowpolicy.Controller{}
}

func InitializeDenyPolicyCommandController(ctx context.Context, param *config.Param) *denypolicy.Controller {
	wire.Build(
		denypolicy.New,
		afero.NewOsFs,
		wire.NewSet(
			policy.NewConfigFinder,
			wire.Bind(new(policy.ConfigFinder), new(*policy.ConfigFinderImpl)),
		),
		wire.NewSet(
			policy.NewValidator,
			wire.Bind(new(policy.Validator), new(*policy.ValidatorImpl)),
		),
	)
	return &denypolicy.Controller{}
}

func InitializeInfoCommandController(ctx context.Context, param *config.Param, rt *runtime.Runtime) *info.Controller {
	wire.Build(
		info.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(info.ConfigFinder), new(*finder.ConfigFinder)),
		),
		afero.NewOsFs,
	)
	return &info.Controller{}
}

func InitializeRemoveCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, httpClient *http.Client, rt *runtime.Runtime, target *config.RemoveMode) *remove.Controller {
	wire.Build(
		remove.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(remove.ConfigFinder), new(*finder.ConfigFinder)),
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(remove.ConfigReader), new(*reader.ConfigReader)),
			wire.Bind(new(which.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(remove.RegistryInstaller), new(*registry.Installer)),
			wire.Bind(new(which.RegistryInstaller), new(*registry.Installer)),
		),
		afero.NewOsFs,
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			osexec.New,
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			fuzzyfinder.New,
			wire.Bind(new(remove.FuzzyFinder), new(*fuzzyfinder.Finder)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(remove.WhichController), new(*which.Controller)),
		),
		osenv.New,
		wire.NewSet(
			link.New,
			wire.Bind(new(installpackage.Linker), new(*link.Linker)),
			wire.Bind(new(which.Linker), new(*link.Linker)),
		),
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(remove.Vacuum), new(*vacuum.Client)),
		),
	)
	return &remove.Controller{}
}

func InitializeVacuumCommandController(ctx context.Context, param *config.Param, rt *runtime.Runtime) *cvacuum.Controller {
	wire.Build(
		cvacuum.New,
		afero.NewOsFs,
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(cvacuum.Vacuum), new(*vacuum.Client)),
		),
	)
	return &cvacuum.Controller{}
}

func InitializeVacuumInitCommandController(ctx context.Context, logE *logrus.Entry, param *config.Param, rt *runtime.Runtime, httpClient *http.Client) *initialize.Controller {
	wire.Build(
		initialize.New,
		afero.NewOsFs,
		wire.NewSet(
			vacuum.New,
			wire.Bind(new(initialize.Vacuum), new(*vacuum.Client)),
		),
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(initialize.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(initialize.ConfigReader), new(*reader.ConfigReader)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(initialize.RegistryInstaller), new(*registry.Installer)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(download.GitHub), new(*github.RepositoriesService)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesService)),
		),
		wire.NewSet(
			github.NewGHES,
			wire.Bind(new(download.GHESResolver), new(*github.GHESRepositoryService)),
			wire.Bind(new(download.GHESContentAPIResolver), new(*github.GHESRepositoryService)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(registry.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(installpackage.CosignVerifier), new(*cosign.Verifier)),
			wire.Bind(new(registry.CosignVerifier), new(*cosign.Verifier)),
		),
		wire.NewSet(
			osexec.New,
			wire.Bind(new(cosign.Executor), new(*osexec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*osexec.Executor)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(installpackage.SLSAVerifier), new(*slsa.Verifier)),
			wire.Bind(new(registry.SLSAVerifier), new(*slsa.Verifier)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return &initialize.Controller{}
}
