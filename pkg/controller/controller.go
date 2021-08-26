package controller

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/go-github/v38/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-template-unmarshaler/text"
	"golang.org/x/oauth2"
)

type Controller struct {
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	ConfigFinder ConfigFinder
	ConfigReader ConfigReader
	GitHub       *github.Client
	RootDir      string
}

func New(ctx context.Context, param *Param) (*Controller, error) {
	if param.LogLevel != "" {
		lvl, err := logrus.ParseLevel(param.LogLevel)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"log_level": param.LogLevel,
			}).WithError(err).Error("the log level is invalid")
		}
		logrus.SetLevel(lvl)
	}
	logrus.WithFields(logrus.Fields{
		"log_level": param.LogLevel,
		"config":    param.ConfigFilePath,
	}).Debug("CLI args")
	text.SetTemplateFunc(func(s string) (*template.Template, error) {
		return template.New("_").Funcs(sprig.TxtFuncMap()).Parse(s) //nolint:wrapcheck
	})
	ctrl := Controller{
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		ConfigFinder: &configFinder{},
		ConfigReader: &configReader{},
		RootDir:      os.Getenv("AQUA_ROOT_DIR"),
	}
	if ctrl.RootDir == "" {
		ctrl.RootDir = filepath.Join(os.Getenv("HOME"), ".aqua")
	}
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		ctrl.GitHub = github.NewClient(
			oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: ghToken},
			)))
	}

	return &ctrl, nil
}
