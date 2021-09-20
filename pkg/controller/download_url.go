package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (downloader *pkgDownloader) downloadFromURL(ctx context.Context, u string, httpClient *http.Client) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request: %w", err)
	}
	if resp.StatusCode >= 400 { //nolint:gomnd
		return resp.Body, errInvalidHTTPStatusCode
	}
	return resp.Body, nil
}
