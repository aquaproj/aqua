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
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/slsa"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Cosign struct {
	installer *Installer
}

func NewCosign(param *config.Param, downloader download.ClientAPI, fs afero.Fs, linker domain.Linker, executor Executor, chkDL download.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, policyChecker domain.PolicyChecker, cosignVerifier cosign.VerifierAPI, slsaVerifier slsa.VerifierAPI) *Cosign {
	return &Cosign{
		installer: &Installer{
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
			Checksum: &registry.Checksum{
				Type:       "github_release",
				Asset:      "cosign_checksums.txt",
				FileFormat: "regexp",
				Algorithm:  "sha256",
				Pattern: &registry.ChecksumPattern{
					Checksum: `^(\b[A-Fa-f0-9]{64}\b)`,
					File:     `^\b[A-Fa-f0-9]{64}\b\s+(\S+)$`,
				},
			},
		},
	}

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

	if err := cos.installer.InstallPackage(ctx, logE, &domain.ParamInstallPackage{
		Checksums: checksum.New(), // Check cosign's checksum but not update aqua-checksums.json
		Pkg:       pkg,
		// PolicyConfigs is nil, so the policy check is skipped
	}); err != nil {
		return err
	}

	return nil
}
