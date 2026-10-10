package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// Requester asks an instance for JSON, which is the one thing every client here does the
// same way: a GET that accepts JSON, a status code to refuse on, and a body to decode.
type Requester struct {
	client *http.Client
}

// NewRequester returns a requester reading over the given HTTP client.
func NewRequester(client *http.Client) *Requester {
	return &Requester{client: client}
}

// GetJSON reads the endpoint and decodes it into dest.
func (r *Requester) GetJSON(ctx context.Context, endpoint string, dest any) error {
	b, err := r.get(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("read the instance: %w", slogerr.With(err, "api_endpoint", endpoint))
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("decode the response body as JSON: %w", slogerr.With(err,
			"api_endpoint", endpoint))
	}
	return nil
}

func (r *Requester) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send a http request: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read a response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if message := Said(b); message != "" {
			return nil, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return b, nil
}
