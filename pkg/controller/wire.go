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
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/exec"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/github"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"

	"github.com/google/wire"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func InitializeListCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *list.Controller {
	wire.Build(
		list.NewController,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(list.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		afero.NewOsFs,
		download.NewHTTPDownloader,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return &list.Controller{}
}

func InitializeGenerateRegistryCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, stdout io.Writer) *genrgst.Controller {
	wire.Build(
		genrgst.NewController,
		wire.NewSet(
			github.New,
			wire.Bind(new(genrgst.RepositoriesService), new(*github.RepositoriesServiceImpl)),
		),
		afero.NewOsFs,
		wire.NewSet(
			output.New,
			wire.Bind(new(genrgst.TestdataOutputter), new(*output.Outputter)),
		),
		wire.NewSet(
			cargo.NewClientImpl,
			wire.Bind(new(cargo.Client), new(*cargo.ClientImpl)),
		),
	)
	return &genrgst.Controller{}
}

func InitializeInitCommandController(ctx context.Context, param *config.Param) *initcmd.Controller {
	wire.Build(
		initcmd.New,
		wire.NewSet(
			github.New,
			wire.Bind(new(initcmd.RepositoriesService), new(*github.RepositoriesServiceImpl)),
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

func InitializeGenerateCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *generate.Controller {
	wire.Build(
		generate.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(generate.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(generate.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(versiongetter.GitHubTagClient), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(versiongetter.GitHubReleaseClient), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
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
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			cargo.NewClientImpl,
			wire.Bind(new(cargo.Client), new(*cargo.ClientImpl)),
			wire.Bind(new(versiongetter.CargoClient), new(*cargo.ClientImpl)),
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
	)
	return &generate.Controller{}
}

func InitializeInstallCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *install.Controller {
	wire.Build(
		install.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(install.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		wire.NewSet(
			installpackage.New,
			wire.Bind(new(installpackage.Installer), new(*installpackage.InstallerImpl)),
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
			wire.Bind(new(domain.Linker), new(*link.Linker)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*exec.Executor)),
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
			wire.Bind(new(unarchive.Unarchiver), new(*unarchive.UnarchiverImpl)),
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
			wire.Bind(new(policy.Reader), new(*policy.ReaderImpl)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
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
	)
	return &install.Controller{}
}

func InitializeWhichCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *which.ControllerImpl {
	wire.Build(
		which.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		osenv.New,
		afero.NewOsFs,
		download.NewHTTPDownloader,
		wire.NewSet(
			link.New,
			wire.Bind(new(domain.Linker), new(*link.Linker)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return nil
}

func InitializeExecCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *cexec.Controller {
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
			wire.Bind(new(installpackage.Installer), new(*installpackage.InstallerImpl)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(which.Controller), new(*which.ControllerImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
			wire.Bind(new(cexec.Executor), new(*exec.Executor)),
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*exec.Executor)),
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
			wire.Bind(new(domain.Linker), new(*link.Linker)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			checksum.NewCalculator,
			wire.Bind(new(installpackage.ChecksumCalculator), new(*checksum.Calculator)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(unarchive.Unarchiver), new(*unarchive.UnarchiverImpl)),
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
			wire.Bind(new(policy.Reader), new(*policy.ReaderImpl)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
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
	)
	return &cexec.Controller{}
}

func InitializeUpdateAquaCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *updateaqua.Controller {
	wire.Build(
		updateaqua.New,
		wire.NewSet(
			afero.NewOsFs,
			wire.Bind(new(installpackage.Cleaner), new(afero.Fs)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(updateaqua.RepositoriesService), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			installpackage.New,
			wire.Bind(new(updateaqua.AquaInstaller), new(*installpackage.InstallerImpl)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*exec.Executor)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(unarchive.Unarchiver), new(*unarchive.UnarchiverImpl)),
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
			wire.Bind(new(domain.Linker), new(*link.Linker)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
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
	)
	return &updateaqua.Controller{}
}

func InitializeCopyCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *cp.Controller {
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
			wire.Bind(new(installpackage.Installer), new(*installpackage.InstallerImpl)),
			wire.Bind(new(cp.PackageInstaller), new(*installpackage.InstallerImpl)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(which.Controller), new(*which.ControllerImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(installpackage.Executor), new(*exec.Executor)),
			wire.Bind(new(cexec.Executor), new(*exec.Executor)),
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(unarchive.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
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
			wire.Bind(new(domain.Linker), new(*link.Linker)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			checksum.NewCalculator,
			wire.Bind(new(installpackage.ChecksumCalculator), new(*checksum.Calculator)),
		),
		wire.NewSet(
			unarchive.New,
			wire.Bind(new(unarchive.Unarchiver), new(*unarchive.UnarchiverImpl)),
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
			wire.Bind(new(policy.Reader), new(*policy.ReaderImpl)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
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
	)
	return &cp.Controller{}
}

func InitializeUpdateChecksumCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *updatechecksum.Controller {
	wire.Build(
		updatechecksum.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(updatechecksum.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		wire.NewSet(
			download.NewChecksumDownloader,
			wire.Bind(new(download.ChecksumDownloader), new(*download.ChecksumDownloaderImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		afero.NewOsFs,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
	)
	return &updatechecksum.Controller{}
}

func InitializeUpdateCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *update.Controller {
	wire.Build(
		update.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(update.ConfigFinder), new(*finder.ConfigFinder)),
			wire.Bind(new(which.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
			wire.Bind(new(update.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(update.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(versiongetter.GitHubTagClient), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(versiongetter.GitHubReleaseClient), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		afero.NewOsFs,
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		wire.NewSet(
			exec.New,
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
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
		wire.NewSet(
			cargo.NewClientImpl,
			wire.Bind(new(cargo.Client), new(*cargo.ClientImpl)),
			wire.Bind(new(versiongetter.CargoClient), new(*cargo.ClientImpl)),
		),
		wire.NewSet(
			which.New,
			wire.Bind(new(which.Controller), new(*which.ControllerImpl)),
		),
		wire.NewSet(
			link.New,
			wire.Bind(new(domain.Linker), new(*link.Linker)),
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

func InitializeRemoveCommandController(ctx context.Context, param *config.Param, httpClient *http.Client, rt *runtime.Runtime) *remove.Controller {
	wire.Build(
		remove.New,
		wire.NewSet(
			finder.NewConfigFinder,
			wire.Bind(new(remove.ConfigFinder), new(*finder.ConfigFinder)),
		),
		wire.NewSet(
			reader.New,
			wire.Bind(new(reader.ConfigReader), new(*reader.ConfigReaderImpl)),
		),
		wire.NewSet(
			registry.New,
			wire.Bind(new(registry.Installer), new(*registry.InstallerImpl)),
		),
		afero.NewOsFs,
		wire.NewSet(
			download.NewGitHubContentFileDownloader,
			wire.Bind(new(domain.GitHubContentFileDownloader), new(*download.GitHubContentFileDownloader)),
		),
		wire.NewSet(
			github.New,
			wire.Bind(new(github.RepositoriesService), new(*github.RepositoriesServiceImpl)),
			wire.Bind(new(download.GitHubContentAPI), new(*github.RepositoriesServiceImpl)),
		),
		wire.NewSet(
			cosign.NewVerifier,
			wire.Bind(new(cosign.Verifier), new(*cosign.VerifierImpl)),
		),
		download.NewHTTPDownloader,
		wire.NewSet(
			exec.New,
			wire.Bind(new(cosign.Executor), new(*exec.Executor)),
			wire.Bind(new(slsa.CommandExecutor), new(*exec.Executor)),
		),
		wire.NewSet(
			download.NewDownloader,
			wire.Bind(new(download.ClientAPI), new(*download.Downloader)),
		),
		wire.NewSet(
			slsa.New,
			wire.Bind(new(slsa.Verifier), new(*slsa.VerifierImpl)),
		),
		wire.NewSet(
			slsa.NewExecutor,
			wire.Bind(new(slsa.Executor), new(*slsa.ExecutorImpl)),
		),
		wire.NewSet(
			fuzzyfinder.New,
			wire.Bind(new(remove.FuzzyFinder), new(*fuzzyfinder.Finder)),
		),
	)
	return &remove.Controller{}
}
