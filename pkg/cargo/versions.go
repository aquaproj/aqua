package cargo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/errors"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type Payload struct {
	Versions []*PayloadVersion `json:"versions"`
}

type PayloadVersion struct {
	Num string `json:"num"`
}

type CratePayload struct {
	Crate *Crate `json:"crate"`
}

type Crate struct {
	Homepage    string `json:"homepage"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
}

type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) ListVersions(ctx context.Context, crate string) ([]string, error) {
	u := fmt.Sprintf("https://crates.io/api/v1/crates/%s/versions", crate)
	versions, _, err := listInstallableVersions(ctx, c.client, u)
	return versions, slogerr.With(err, "url", u) //nolint:wrapcheck
}

func (c *Client) GetLatestVersion(ctx context.Context, crate string) (string, error) {
	versions, err := c.ListVersions(ctx, crate)
	if len(versions) == 0 {
		return "", err
	}
	return versions[0], err
}

func (c *Client) GetCrate(ctx context.Context, crate string) (*CratePayload, error) {
	payload, _, err := getCrate(ctx, c.client, "https://crates.io/api/v1/crates/"+crate)
	return payload, err
}

func getCrate(ctx context.Context, client *http.Client, uri string) (*CratePayload, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create a HTTP request: %w", err)
	}
	req.Header.Add("User-Agent", "aqua") // https://github.com/aquaproj/aqua/issues/2742
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { //nolint:mnd
		return nil, resp.StatusCode, errors.ErrHTTPStatusCodeIsGreaterEqualThan300
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
	req.Header.Add("User-Agent", "aqua") // https://github.com/aquaproj/aqua/issues/2742
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 { //nolint:mnd
		return nil, resp.StatusCode, errors.ErrHTTPStatusCodeIsGreaterEqualThan300
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
