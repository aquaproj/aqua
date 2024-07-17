package verify

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/sirupsen/logrus"
)

type Mock struct{}

func (m *Mock) Verify(ctx context.Context, logE *logrus.Entry, pkg *config.Package, bodyFile *download.DownloadedFile) error {
	return nil
}
