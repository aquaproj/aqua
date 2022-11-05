package cp

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

type Installer interface {
	Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error
}

type MockInstaller struct {
	Err error
}

func (inst *MockInstaller) Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	return inst.Err
}
