package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (ctrl *Controller) downloadFromURL(ctx context.Context, u string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request: %w", err)
	}
	if resp.StatusCode >= 400 { //nolint:gomnd
		return resp.Body, errors.New("status code >= 400")
	}
	return resp.Body, nil
}
