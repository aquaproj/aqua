package domain

import (
	"context"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

type ChecksumDownloader interface {
	DownloadChecksum(ctx context.Context, logE *logrus.Entry, pkg *config.Package) (io.ReadCloser, int64, error)
}
