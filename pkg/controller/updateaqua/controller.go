package updateaqua

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

const filePermission os.FileMode = 0o755

type Controller struct {
	rootDir     string
	fs          afero.Fs
	runtime     *runtime.Runtime
	execFinder  domain.ExecFinder
	github      RepositoriesService
	ghReleaseDL domain.GitHubReleaseDownloader
}

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, execFinder domain.ExecFinder, ghReleaseDL domain.GitHubReleaseDownloader, gh RepositoriesService) *Controller {
	return &Controller{
		rootDir:     param.RootDir,
		execFinder:  execFinder,
		fs:          fs,
		runtime:     rt,
		ghReleaseDL: ghReleaseDL,
		github:      gh,
	}
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

type ExecFinder struct{}

func (finder *ExecFinder) LookPath(p string) (string, error) {
	return exec.LookPath(p) //nolint:wrapcheck
}

func NewExecFinder() *ExecFinder {
	return &ExecFinder{}
}

func (ctrl *Controller) UpdateAqua(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	p, err := ctrl.execFinder.LookPath("aqua")
	if err != nil {
		return fmt.Errorf("look up aqua: %w", err)
	}

	release, _, err := ctrl.github.GetLatestRelease(ctx, "aquaproj", "aqua")
	if err != nil {
		return fmt.Errorf("get the latest version of aqua: %w", err)
	}
	latestAquaVersion := release.GetTagName()
	logE = logE.WithField("new_version", latestAquaVersion)

	if param.AQUAVersion == latestAquaVersion {
		logE.Info("aqua is already latest")
		return nil
	}

	assetName := fmt.Sprintf("aqua_%s_%s.tar.gz", ctrl.runtime.GOOS, ctrl.runtime.GOARCH)
	file, _, err := ctrl.ghReleaseDL.DownloadGitHubRelease(ctx, logE, &domain.DownloadGitHubReleaseParam{
		RepoOwner: "aquaproj",
		RepoName:  "aqua",
		Version:   latestAquaVersion,
		Asset:     assetName,
	})
	if err != nil {
		return fmt.Errorf("download aqua: %w", logerr.WithFields(err, logrus.Fields{
			"new_version": latestAquaVersion,
			"asset":       assetName,
		}))
	}
	defer file.Close()
	logE.Info("updating aqua")
	// dest, err := ctrl.fs.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission)
	// if err != nil {
	// 	return fmt.Errorf("create a file: %w", err)
	// }
	dest, err := ctrl.fs.Create(p)
	if err != nil {
		return fmt.Errorf("create a file: %w", err)
	}
	defer dest.Close()
	if err := ctrl.fs.Chmod(p, filePermission); err != nil {
		return fmt.Errorf("change a file permission: %w", err)
	}
	if err := unarchive(dest, file); err != nil {
		return fmt.Errorf("downloand and unarchive aqua: %w", err)
	}
	return nil
}
