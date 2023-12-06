package download

import (
	"context"
	"io"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/sirupsen/logrus"
)

type MockChecksumDownloader struct {
	Body string
	Code int64
	Err  error
}

func (dl *MockChecksumDownloader) DownloadChecksum(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error) {
	return io.NopCloser(strings.NewReader(dl.Body)), dl.Code, dl.Err
}

type Mock struct {
	RC   io.ReadCloser
	Code int64
	Err  error
}

func (m *Mock) ReadCloser(ctx context.Context, logE *logrus.Entry, file *File) (io.ReadCloser, int64, error) {
	return m.RC, m.Code, m.Err
}
