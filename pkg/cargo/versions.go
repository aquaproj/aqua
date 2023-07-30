package cargo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/util"
)

type Payload struct {
	Versions []*PayloadVersion `json:"versions"`
}

type PayloadVersion struct {
	Num string `json:"num"`
}

type Client interface {
	ListVersions(ctx context.Context, crate string) ([]string, error)
	GetLatestVersion(ctx context.Context, crate string) (string, error)
	GetCrate(ctx context.Context, crate string) (*CratePayload, error)
}

type CratePayload struct {
	Crate *Crate `json:"crate"`
}

type Crate struct {
	Homepage    string `json:"homepage"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
}

type MockClient struct {
	Versions     []string
	Err          error
	CratePayload *CratePayload
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

func (mock *MockClient) GetCrate(ctx context.Context, crate string) (*CratePayload, error) {
	return mock.CratePayload, mock.Err
}

type ClientImpl struct {
	client *http.Client
}

func NewClientImpl(client *http.Client) *ClientImpl {
	return &ClientImpl{
		client: client,
	}
}

func (searcher *ClientImpl) ListVersions(ctx context.Context, crate string) ([]string, error) {
	versions, _, err := listInstallableVersions(ctx, searcher.client, fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", crate))
	return versions, err
}

func (searcher *ClientImpl) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	versions, err := searcher.ListVersions(ctx, crate)
	if len(versions) == 0 {
		return "", err
	}
	return versions[0], err
}

func (searcher *ClientImpl) GetCrate(ctx context.Context, crate string) (*CratePayload, error) {
	payload, _, err := getCrate(ctx, searcher.client, fmt.Sprintf("https://crates.io/api/v1/crates/%s", crate))
	return payload, err
}

func getCrate(ctx context.Context, client *http.Client, uri string) (*CratePayload, int, error) {
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
	payload := &CratePayload{}
	if err := json.NewDecoder(resp.Body).Decode(payload); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode the response body as JSON: %w", err)
	}
	return payload, resp.StatusCode, nil
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
	versions := make([]string, len(payload.Versions))
	for i, v := range payload.Versions {
		versions[i] = v.Num
	}
	return versions, resp.StatusCode, nil
}
