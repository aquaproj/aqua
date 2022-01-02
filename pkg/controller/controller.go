package controller

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/log"
	"github.com/google/go-github/v39/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Controller struct {
	Stdin                   io.Reader
	Stdout                  io.Writer
	Stderr                  io.Writer
	ConfigFinder            ConfigFinder
	ConfigReader            ConfigReader
	GitHubRepositoryService GitHubRepositoryService
	PackageDownloader       PackageDownloader
	RootDir                 string
	Version                 string
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

func New(ctx context.Context, param *Param) (*Controller, error) {
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
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		ConfigFinder: &configFinder{},
		ConfigReader: &configReader{},
		RootDir:      os.Getenv("AQUA_ROOT_DIR"),
		Version:      param.AQUAVersion,
	}
	if ctrl.RootDir == "" {
		ctrl.RootDir = filepath.Join(os.Getenv("HOME"), ".aqua")
	}
	ctrl.GitHubRepositoryService = github.NewClient(getHTTPClientForGitHub(ctx, getGitHubToken())).Repositories
	ctrl.PackageDownloader = &pkgDownloader{
		GitHubRepositoryService: ctrl.GitHubRepositoryService,
		logE:                    ctrl.logE,
	}

	return &ctrl, nil
}

func (ctrl *Controller) logE() *logrus.Entry {
	return log.New().WithField("aqua_version", ctrl.Version)
}
