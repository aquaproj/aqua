package download

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type MockChecksumDownloader struct {
	Body string
	Code int64
	Err  error
}

func (dl *MockChecksumDownloader) DownloadChecksum(ctx context.Context, logger *slog.Logger, rt *runtime.Runtime, pkg *config.Package) (io.ReadCloser, int64, error) {
	return io.NopCloser(strings.NewReader(dl.Body)), dl.Code, dl.Err
}

func (dl *MockChecksumDownloader) GetReleaseAssets(ctx context.Context, logger *slog.Logger, pkg *config.Package) (domain.ReleaseAssets, error) {
	return nil, nil //nolint:nilnil
}

type Mock struct {
	RC   io.ReadCloser
	Code int64
	Err  error
}

func (m *Mock) ReadCloser(ctx context.Context, logger *slog.Logger, file *File) (io.ReadCloser, int64, error) {
	return m.RC, m.Code, m.Err
}
