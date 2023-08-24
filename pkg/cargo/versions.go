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

func (m *MockClient) ListVersions(ctx context.Context, crate string) ([]string, error) {
	return m.Versions, m.Err
}

func (m *MockClient) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	if len(m.Versions) == 0 {
		return "", m.Err
	}
	return m.Versions[0], m.Err
}

func (m *MockClient) GetCrate(ctx context.Context, crate string) (*CratePayload, error) {
	return m.CratePayload, m.Err
}

type ClientImpl struct {
	client *http.Client
}

func NewClientImpl(client *http.Client) *ClientImpl {
	return &ClientImpl{
		client: client,
	}
}

func (c *ClientImpl) ListVersions(ctx context.Context, crate string) ([]string, error) {
	versions, _, err := listInstallableVersions(ctx, c.client, fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", crate))
	return versions, err
}

func (c *ClientImpl) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	versions, err := c.ListVersions(ctx, crate)
	if len(versions) == 0 {
		return "", err
	}
	return versions[0], err
}

func (c *ClientImpl) GetCrate(ctx context.Context, crate string) (*CratePayload, error) {
	payload, _, err := getCrate(ctx, c.client, fmt.Sprintf("https://crates.io/api/v1/crates/%s", crate))
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
