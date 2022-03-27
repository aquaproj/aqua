package controller

import (
	"io"
	"os"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	githubSvc "github.com/aquaproj/aqua/pkg/github"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/installpackage"
	"github.com/aquaproj/aqua/pkg/log"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	stdin                   io.Reader
	stdout                  io.Writer
	stderr                  io.Writer
	configFinder            finder.ConfigFinder
	configReader            reader.ConfigReader
	gitHubRepositoryService githubSvc.RepositoryService
	registryInstaller       registry.Installer
	packageInstaller        installpackage.Installer
	rootDir                 string
	logger                  *log.Logger
}

func New(rootDir config.RootDir, configFinder finder.ConfigFinder, configReader reader.ConfigReader, logger *log.Logger, pkgInstaller installpackage.Installer, gh githubSvc.RepositoryService, registInstaller registry.Installer, param *config.Param) (*Controller, error) {
	if param.LogLevel != "" {
		lvl, err := logrus.ParseLevel(param.LogLevel)
		if err != nil {
			logger.LogE().WithField("log_level", param.LogLevel).WithError(err).Error("the log level is invalid")
		}
		logrus.SetLevel(lvl)
	}
	logger.LogE().WithFields(logrus.Fields{
		"log_level": param.LogLevel,
		"config":    param.ConfigFilePath,
	}).Debug("CLI args")
	ctrl := Controller{
		stdin:                   os.Stdin,
		stdout:                  os.Stdout,
		stderr:                  os.Stderr,
		configFinder:            configFinder,
		configReader:            configReader,
		rootDir:                 string(rootDir),
		gitHubRepositoryService: gh,
		packageInstaller:        pkgInstaller,
		registryInstaller:       registInstaller,
		logger:                  logger,
	}

	return &ctrl, nil
}

func (ctrl *Controller) logE() *logrus.Entry {
	return ctrl.logger.LogE()
}
