package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPDownloader interface {
	Download(ctx context.Context, u string) (io.ReadCloser, error)
}

func NewHTTPDownloader(httpClient *http.Client) HTTPDownloader {
	return &httpDownloader{
		client: httpClient,
	}
}

type httpDownloader struct {
	client *http.Client
}

func (downloader *httpDownloader) Download(ctx context.Context, u string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	resp, err := downloader.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request: %w", err)
	}
	if resp.StatusCode >= 400 { //nolint:gomnd
		return resp.Body, errInvalidHTTPStatusCode
	}
	return resp.Body, nil
}
