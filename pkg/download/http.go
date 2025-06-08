package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type HTTPDownloader interface {
	Download(ctx context.Context, logE *logrus.Entry, u string) (io.ReadCloser, int64, error)
}

func NewHTTPDownloader(httpClient *http.Client) HTTPDownloader {
	return &httpDownloader{
		client: httpClient,
	}
}

type httpDownloader struct {
	client *http.Client
}

func (dl *httpDownloader) Download(ctx context.Context, logE *logrus.Entry, u string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create a http request: %w", err)
	}

	const retryDelay = 100 * time.Millisecond

	for cnt := range 5 {
		resp, err := dl.client.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("send http request: %w", err)
		}

		if resp.StatusCode < http.StatusBadRequest {
			// Success case
			return resp.Body, resp.ContentLength, nil
		}
		resp.Body.Close()

		if cnt != 4 && resp.StatusCode >= http.StatusInternalServerError {
			// wait and retry
			logE.WithFields(logrus.Fields{
				"http_status_code": resp.StatusCode,
			}).Warn("downloading a file failed. Retrying...")
			if err := timer.Wait(ctx, retryDelay); err != nil {
				// Context was canceled during wait
				return nil, 0, fmt.Errorf("context canceled while waiting to retry: %w", err)
			}
			continue
		}

		return nil, 0, logerr.WithFields(errInvalidHTTPStatusCode, logrus.Fields{ //nolint:wrapcheck
			"http_status_code": resp.StatusCode,
		})
	}

	// This should not be reached, but just in case
	return nil, 0, errors.New("failed to download a file")
}
