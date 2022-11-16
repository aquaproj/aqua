package installpackage

import (
	"context"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/spf13/afero"
)

type Executor interface {
	GoInstall(ctx context.Context, path, gobin string) (int, error)
}

func New(param *config.Param, downloader domain.PackageDownloader, rt *runtime.Runtime, fs afero.Fs, linker domain.Linker, executor Executor, chkDL domain.ChecksumDownloader, chkCalc ChecksumCalculator, unarchiver Unarchiver, policyChecker domain.PolicyChecker) *Installer {
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
	}
}
