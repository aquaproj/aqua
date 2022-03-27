package download_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/download"
	"github.com/suzuki-shunsuke/flute/flute"
)

func TestFromURL(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title      string
		url        string
		httpClient *http.Client
		isErr      bool
		body       string
	}{
		{
			title: "normal",
			url:   "http://example.com/v0.1.0/foo",
			body:  "xxxxxx",
			httpClient: &http.Client{
				Transport: &flute.Transport{
					Services: []flute.Service{
						{
							Endpoint: "http://example.com",
							Routes: []flute.Route{
								{
									Name: "download an asset",
									Matcher: &flute.Matcher{
										Method: "GET",
										Path:   "/v0.1.0/foo",
									},
									Response: &flute.Response{
										Base: http.Response{
											StatusCode: 200,
										},
										BodyString: "xxxxxx",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			readCloser, err := download.FromURL(ctx, d.url, d.httpClient)
			if readCloser != nil {
				defer readCloser.Close()
			}
			if d.isErr {
				if err != nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			b, err := io.ReadAll(readCloser)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != d.body {
				t.Fatalf("wanted %s, got %s", d.body, string(b))
			}
		})
	}
}
