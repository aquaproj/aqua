package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type HTTPDownloader interface {
	Download(ctx context.Context, u string) (io.ReadCloser, int64, error)
}

func NewHTTPDownloader(httpClient *http.Client) HTTPDownloader {
	return &httpDownloader{
		client: httpClient,
	}
}

type httpDownloader struct {
	client *http.Client
}

func (dl *httpDownloader) Download(ctx context.Context, u string) (io.ReadCloser, int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create a http request: %w", err)
	}

	const maxRetries = 4
	const retryDelay = 100 * time.Millisecond

	var resp *http.Response
	var retryCount int

	for retryCount = 0; retryCount <= maxRetries; retryCount++ {
		if retryCount > 0 {
			if err := timer.Wait(ctx, retryDelay); err != nil {
				// Context was canceled during wait
				return nil, 0, fmt.Errorf("context canceled while waiting to retry: %w", err)
			}
		}

		resp, err = dl.client.Do(req)
		if err != nil {
			if retryCount == maxRetries {
				return nil, 0, fmt.Errorf("send http request: %w", err)
			}
			continue
		}

		// For 4xx errors, don't retry
		if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
			return resp.Body, 0, logerr.WithFields(errInvalidHTTPStatusCode, logrus.Fields{ //nolint:wrapcheck
				"http_status_code": resp.StatusCode,
			})
		}

		// For 5xx errors, retry if we haven't reached max retries
		if resp.StatusCode >= http.StatusInternalServerError {
			if retryCount < maxRetries {
				resp.Body.Close() // Close the body before retrying
				continue
			}
			// If we've reached max retries, return the error
			return resp.Body, 0, logerr.WithFields(errInvalidHTTPStatusCode, logrus.Fields{ //nolint:wrapcheck
				"http_status_code": resp.StatusCode,
				"retries":          retryCount,
			})
		}

		// Success case
		return resp.Body, resp.ContentLength, nil
	}

	// This should not be reached, but just in case
	return nil, 0, fmt.Errorf("failed to download after %d retries", maxRetries)
}
