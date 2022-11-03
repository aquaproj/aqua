package updateaqua

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type Controller struct {
	rootDir   string
	fs        afero.Fs
	runtime   *runtime.Runtime
	github    RepositoriesService
	installer AquaInstaller
}

type AquaInstaller interface {
	InstallAqua(ctx context.Context, logE *logrus.Entry, version string) error
}

func New(param *config.Param, fs afero.Fs, rt *runtime.Runtime, gh RepositoriesService, installer AquaInstaller) *Controller {
	return &Controller{
		rootDir:   param.RootDir,
		fs:        fs,
		runtime:   rt,
		github:    gh,
		installer: installer,
	}
}

type RepositoriesService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

const dirPermission os.FileMode = 0o775

func (ctrl *Controller) UpdateAqua(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	rootBin := filepath.Join(ctrl.rootDir, "bin")
	if err := ctrl.fs.MkdirAll(rootBin, dirPermission); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	if ctrl.runtime.GOOS == "windows" {
		if err := ctrl.fs.MkdirAll(filepath.Join(ctrl.rootDir, "bat"), dirPermission); err != nil {
			return fmt.Errorf("create the directory: %w", err)
		}
	}

	release, _, err := ctrl.github.GetLatestRelease(ctx, "aquaproj", "aqua")
	if err != nil {
		return fmt.Errorf("get the latest version of aqua: %w", err)
	}
	latestAquaVersion := release.GetTagName()
	logE = logE.WithField("new_version", latestAquaVersion)

	if err := ctrl.installer.InstallAqua(ctx, logE, latestAquaVersion); err != nil {
		return fmt.Errorf("download aqua: %w", logerr.WithFields(err, logrus.Fields{
			"new_version": latestAquaVersion,
		}))
	}
	return nil
}
