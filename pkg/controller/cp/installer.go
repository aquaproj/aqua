package cp

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
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
