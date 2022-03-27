package controller

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/download"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/google/go-github/v39/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Controller struct {
	stdin                   io.Reader
	stdout                  io.Writer
	stderr                  io.Writer
	configFinder            ConfigFinder
	configReader            ConfigReader
	gitHubRepositoryService GitHubRepositoryService
	registryInstaller       registry.Installer
	packageInstaller        installpackage.Installer
	rootDir                 string
	globalConfingDir        string
	version                 string
}

type GitHubRepositoryService interface {
	GetLatestRelease(ctx context.Context, repoOwner, repoName string) (*github.RepositoryRelease, *github.Response, error)
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}

func getGitHubToken() string {
	if token := os.Getenv("AQUA_GITHUB_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("GITHUB_TOKEN")
}

func getHTTPClientForGitHub(ctx context.Context, token string) *http.Client {
	if token == "" {
		return http.DefaultClient
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))
}

func New(ctx context.Context, param *config.Param) (*Controller, error) {
	if param.LogLevel != "" {
		lvl, err := logrus.ParseLevel(param.LogLevel)
		if err != nil {
			log.New().WithFields(logrus.Fields{
				"log_level":    param.LogLevel,
				"aqua_version": param.AQUAVersion,
			}).WithError(err).Error("the log level is invalid")
		}
		logrus.SetLevel(lvl)
	}
	log.New().WithFields(logrus.Fields{
		"log_level":    param.LogLevel,
		"config":       param.ConfigFilePath,
		"aqua_version": param.AQUAVersion,
	}).Debug("CLI args")
	ctrl := Controller{
		stdin:        os.Stdin,
		stdout:       os.Stdout,
		stderr:       os.Stderr,
		configFinder: &finder.ConfigFinder{},
		configReader: &configReader{
			reader: &fileReader{},
		},
		rootDir: os.Getenv("AQUA_ROOT_DIR"),
		version: param.AQUAVersion,
	}
	if ctrl.rootDir == "" {
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			xdgDataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
		}
		ctrl.rootDir = filepath.Join(xdgDataHome, "aquaproj-aqua")
	}
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	ctrl.globalConfingDir = filepath.Join(xdgConfigHome, "aquaproj-aqua")

	ctrl.gitHubRepositoryService = github.NewClient(getHTTPClientForGitHub(ctx, getGitHubToken())).Repositories
	ctrl.packageInstaller = installpackage.New(ctrl.version, ctrl.rootDir, &download.PkgDownloader{
		GitHubRepositoryService: ctrl.gitHubRepositoryService,
		LogE:                    ctrl.logE,
	})
	ctrl.registryInstaller = registry.New(ctrl.version, ctrl.rootDir, download.NewRegistryDownloader(ctrl.gitHubRepositoryService, ctrl.version))

	return &ctrl, nil
}

func (ctrl *Controller) logE() *logrus.Entry {
	return log.New().WithField("aqua_version", ctrl.version)
}
