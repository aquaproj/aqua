package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Executor interface {
	GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error)
	GoInstall(ctx context.Context, path, gobin string) (int, error)
}

type CosignVerifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error
	HasCosign() bool
}

func New(param *config.Param, downloader domain.PackageDownloader, rt *runtime.Runtime, fs afero.Fs, linker domain.Linker, executor Executor, chkDL domain.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, policyChecker domain.PolicyChecker, cosignVerifier CosignVerifier) *Installer {
	return &Installer{
		rootDir:            param.RootDir,
		maxParallelism:     param.MaxParallelism,
		packageDownloader:  downloader,
		checksumDownloader: chkDL,
		checksumFileParser: &checksum.FileParser{},
		checksumCalculator: chkCalc,
		runtime:            rt,
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
	}
}
