package pypi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/aquaproj/aqua/v2/pkg/util"
)

type Payload struct {
	Releases map[string]interface{} `json:"releases"`
}

type PayloadVersion struct {
	Num string `json:"num"`
}

type Client interface {
	ListVersions(ctx context.Context, crate string) ([]string, error)
	GetLatestVersion(ctx context.Context, crate string) (string, error)
}

type MockClient struct {
	Versions []string
	Err      error
}

func (mock *MockClient) ListVersions(ctx context.Context, crate string) ([]string, error) {
	return mock.Versions, mock.Err
}

func (mock *MockClient) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	if len(mock.Versions) == 0 {
		return "", mock.Err
	}
	return mock.Versions[0], mock.Err
}

type ClientImpl struct {
	client *http.Client
}

func NewClientImpl(client *http.Client) *ClientImpl {
	return &ClientImpl{
		client: client,
	}
}

func (searcher *ClientImpl) ListVersions(ctx context.Context, pkgName string) ([]string, error) {
	versions, _, err := listInstallableVersions(ctx, searcher.client, fmt.Sprintf("https://pypi.org/pypi/%s/json", pkgName))
	return versions, err
}

func (searcher *ClientImpl) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	versions, err := searcher.ListVersions(ctx, crate)
	if len(versions) == 0 {
		return "", err
	}
	return versions[len(versions)-1], err
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
		return nil, resp.StatusCode, util.ErrHTTPStatusCodeIsGreaterEqualThan300
	}
	payload := &Payload{}
	if err := json.NewDecoder(resp.Body).Decode(payload); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode the response body as JSON: %w", err)
	}
	versions := make([]string, 0, len(payload.Releases))
	for v := range payload.Releases {
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})
	return versions, resp.StatusCode, nil
}
