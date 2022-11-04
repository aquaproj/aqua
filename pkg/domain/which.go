package domain

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type WhichController interface {
	Which(ctx context.Context, param *config.Param, exeName string, logE *logrus.Entry) (*FindResult, error)
}

type MockWhichController struct {
	FindResult *FindResult
	Err        error
}

func (ctrl *MockWhichController) Which(ctx context.Context, param *config.Param, exeName string, logE *logrus.Entry) (*FindResult, error) {
	return ctrl.FindResult, ctrl.Err
}

type FindResult struct {
	Package        *config.Package
	File           *registry.File
	Config         *aqua.Config
	ExePath        string
	ConfigFilePath string
	EnableChecksum bool
}
