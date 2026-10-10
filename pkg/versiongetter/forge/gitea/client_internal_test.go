package gitea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClient_ListReleases(t *testing.T) { //nolint:cyclop
	t.Parallel()
	var got *url.URL
	var accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL
		accept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		// Trimmed to the keys aqua reads; an instance says a good deal more, which
		// has to be ignored rather than refused.
		_, _ = w.Write([]byte(`[
		  {
		    "id": 1,
		    "tag_name": "v0.20.0",
		    "name": "v0.20.0",
		    "body": "the notes",
		    "html_url": "https://codeberg.org/mergiraf/mergiraf/releases/tag/v0.20.0",
		    "draft": false,
		    "prerelease": false,
		    "hide_archive_links": false,
		    "assets": [{"name": "mergiraf_x86_64-unknown-linux-gnu.tar.gz", "uuid": "an-uuid"}]
		  }
		]`))
	}))
	defer server.Close()

	client := New(toServer(server))

	releases, err := client.ListReleases(context.Background(), "codeberg.org", "mergiraf", "mergiraf", 2, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) != 1 {
		t.Fatalf("wanted 1 release, got %d", len(releases))
	}
	release := releases[0]
	if release.TagName != "v0.20.0" {
		t.Fatalf("wanted the tag, got %s", release.TagName)
	}
	if release.Name != "v0.20.0" || release.Body != "the notes" || release.HTMLURL == "" {
		t.Fatalf("wanted what the preview is written from, got %+v", release)
	}
	if release.Draft || release.Prerelease {
		t.Fatalf("wanted a release that is neither a draft nor a prerelease, got %+v", release)
	}
	if got.Path != "/api/v1/repos/mergiraf/mergiraf/releases" {
		t.Fatalf("wanted the v1 releases endpoint, got %s", got.Path)
	}
	// page and limit rather than per_page: this is Gitea's API rather than GitHub's.
	if q := got.Query(); q.Get("page") != "2" || q.Get("limit") != "30" {
		t.Fatalf("wanted page 2 of 30, got %s", got.RawQuery)
	}
	if accept != "application/json" {
		t.Fatalf("wanted a JSON request, got %s", accept)
	}
}

func TestClient_ListReleases_notFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"The target couldn't be found.","url":"https://codeberg.org/api/swagger"}`))
	}))
	defer server.Close()

	client := New(toServer(server))
	_, err := client.ListReleases(context.Background(), "codeberg.org", "an-owner", "a-repo", 1, 30)
	if err == nil {
		t.Fatal("an error must be returned")
	}
	// What the instance said about it, which is where a rate limit, a renamed
	// repository and a private one are told apart.
	if !strings.Contains(err.Error(), "The target couldn't be found.") {
		t.Fatalf("wanted the error to carry what the instance said, got %v", err)
	}
}

// toServer is an HTTP client that answers from the test server whatever instance it is
// asked for, so that the request the client builds is the one under test rather than
// something the test had to tell it to build.
func toServer(server *httptest.Server) *http.Client {
	u, err := url.Parse(server.URL)
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
			to := req.Clone(req.Context())
			to.URL.Scheme = u.Scheme
			to.URL.Host = u.Host
			return server.Client().Transport.RoundTrip(to)
		}),
	}
}

type roundTripper func(req *http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
