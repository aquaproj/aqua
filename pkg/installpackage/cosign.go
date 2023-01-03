package installpackage

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/policy"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/slsa"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Cosign struct {
	installer *InstallerImpl
}

func NewCosign(param *config.Param, downloader download.ClientAPI, fs afero.Fs, linker domain.Linker, executor Executor, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, policyChecker policy.Checker, cosignVerifier cosign.Verifier, slsaVerifier slsa.Verifier) *Cosign {
	return &Cosign{
		installer: &InstallerImpl{
			rootDir:            param.RootDir,
			maxParallelism:     param.MaxParallelism,
			downloader:         downloader,
			checksumDownloader: chkDL,
			checksumFileParser: &checksum.FileParser{},
			checksumCalculator: chkCalc,
			runtime:            runtime.NewR(),
			fs:                 fs,
			linker:             linker,
			executor:           executor,
			progressBar:        param.ProgressBar,
			isTest:             param.IsTest,
			onlyLink:           param.OnlyLink,
			copyDir:            param.Dest,
			unarchiver:         unarchiver,
			policyChecker:      policyChecker,
			cosign:             cosignVerifier,
			slsaVerifier:       slsaVerifier,
		},
	}
}

func (cos *Cosign) InstallCosign(ctx context.Context, logE *logrus.Entry, version string) error {
	assetTemplate := `cosign-{{.OS}}-{{.Arch}}`
	pkg := &config.Package{
		Package: &aqua.Package{
			Name:    "sigstore/cosign",
			Version: version,
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "sigstore",
			RepoName:  "cosign",
			Asset:     &assetTemplate,
			SupportedEnvs: []string{
				"darwin",
				"linux",
				"amd64",
			},
		},
	}

	checksums := map[string]string{
		"darwin/amd64":  "1d164b8b1fcfef1e1870d809edbb9862afd5995cab63687a440b84cca5680ecf",
		"darwin/arm64":  "02bef878916be048fd7dcf742105639f53706a59b5b03f4e4eaccc01d05bc7ab",
		"linux/amd64":   "a50651a67b42714d6f1a66eb6773bf214dacae321f04323c0885f6a433051f95",
		"linux/arm64":   "a7a79a52c7747e2c21554cad4600e6c7130c0429017dd258f9c558d957fa9090",
		"windows/amd64": "78a2774b68b995cc698944f6c235b1c93dcb6d57593a58a565ee7a56d64e4b85",
	}
	chksum := checksums[cos.installer.runtime.Env()]

	pkgInfo, err := pkg.PackageInfo.Override(pkg.Package.Version, cos.installer.runtime)
	if err != nil {
		return fmt.Errorf("evaluate version constraints: %w", err)
	}
	supported, err := pkgInfo.CheckSupported(cos.installer.runtime, cos.installer.runtime.GOOS+"/"+cos.installer.runtime.GOARCH)
	if err != nil {
		return fmt.Errorf("check if cosign is supported: %w", err)
	}
	if !supported {
		logE.Debug("the package isn't supported on this environment")
		return nil
	}

	pkg.PackageInfo = pkgInfo

	if err := cos.installer.InstallPackage(ctx, logE, &ParamInstallPackage{
		Checksums: checksum.New(), // Check cosign's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		Checksum: &checksum.Checksum{
			Algorithm: "sha256",
			Checksum:  chksum,
		},
		// PolicyConfigs is nil, so the policy check is skipped
	}); err != nil {
		return err
	}

	return nil
}
