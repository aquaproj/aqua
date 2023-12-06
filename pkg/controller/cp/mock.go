package cp

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

type MockInstaller struct {
	Err error
}

func (is *MockInstaller) Install(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	return is.Err
}
