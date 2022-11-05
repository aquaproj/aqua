package domain

import (
	"context"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type ChecksumDownloader interface {
	DownloadChecksum(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error)
}

type MockChecksumDownloader struct {
	Body string
	Code int64
	Err  error
}

func (chkDL *MockChecksumDownloader) DownloadChecksum(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error) {
	return io.NopCloser(strings.NewReader(chkDL.Body)), chkDL.Code, chkDL.Err
}
