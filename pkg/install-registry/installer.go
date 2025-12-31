package registry

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/spf13/afero"
)

type Installer struct {
	registryDownloader GitHubContentFileDownloader
	param              *config.Param
	fs                 afero.Fs
	cosign             CosignVerifier
	slsaVerifier       SLSAVerifier
	rt                 *runtime.Runtime
}

func New(param *config.Param, downloader GitHubContentFileDownloader, fs afero.Fs, rt *runtime.Runtime, cos CosignVerifier, slsaVerifier SLSAVerifier) *Installer {
	return &Installer{
		param:              param,
		registryDownloader: downloader,
		fs:                 fs,
		rt:                 rt,
		cosign:             cos,
		slsaVerifier:       slsaVerifier,
	}
}

type GitHubContentFileDownloader interface {
	DownloadGitHubContentFile(ctx context.Context, logger *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error)
}

type SLSAVerifier interface {
	Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *slsa.ParamVerify) error
}

type CosignVerifier interface {
	Verify(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, file *download.File, cos *registry.Cosign, art *template.Artifact, verifiedFilePath string) error
}
