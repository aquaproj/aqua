package cargo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Payload struct {
	Versions []*PayloadVersion `json:"versions"`
}

type PayloadVersion struct {
	Num string `json:"num"`
}

type VersionSearcher interface {
	List(ctx context.Context, crate string) ([]string, error)
	GetLatest(ctx context.Context, crate string) (string, error)
}

type MockVersionSearcher struct {
	Versions []string
	Err      error
}

func (mock *MockVersionSearcher) List(ctx context.Context, crate string) ([]string, error) {
	return mock.Versions, mock.Err
}

func (mock *MockVersionSearcher) GetLatest(ctx context.Context, crate string) (string, error) {
	if len(mock.Versions) == 0 {
		return "", mock.Err
	}
	return mock.Versions[0], mock.Err
}

type VersionSearcherImpl struct {
	client *http.Client
}

func NewVersionSearcherImpl(client *http.Client) *VersionSearcherImpl {
	return &VersionSearcherImpl{
		client: client,
	}
}

func (searcher *VersionSearcherImpl) List(ctx context.Context, crate string) ([]string, error) {
	versions, _, err := listInstallableVersions(ctx, searcher.client, fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", crate))
	return versions, err
}

func (searcher *VersionSearcherImpl) GetLatest(ctx context.Context, crate string) (string, error) {
	versions, err := searcher.List(ctx, crate)
	if len(versions) == 0 {
		return "", err
	}
	return versions[0], err
}

func listInstallableVersions(ctx context.Context, client *http.Client, uri string) ([]string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create a HTTP request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { //nolint:gomnd
		return nil, resp.StatusCode, errHTTPStatusCodeIsGreaterEqualThan300
	}
	payload := &Payload{}
	if err := json.NewDecoder(resp.Body).Decode(payload); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode the response body as JSON: %w", err)
	}
	versions := make([]string, len(payload.Versions))
	for i, v := range payload.Versions {
		versions[i] = v.Num
	}
	return versions, resp.StatusCode, nil
}
