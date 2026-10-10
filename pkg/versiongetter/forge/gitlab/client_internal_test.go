package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// GitLab answers in its own words, and the translation is what the version getter reads,
// so this is about the request being the one GitLab wants and the fields it answers with
// meaning what the getter thinks they mean.
func TestClient_ListReleases(t *testing.T) { //nolint:cyclop
	t.Parallel()
	var got *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL
		w.Header().Set("Content-Type", "application/json")
		// Trimmed to the keys aqua reads; a release says a good deal more, which has
		// to be ignored rather than refused.
		_, _ = w.Write([]byte(`[
		  {
		    "name": "v1.122.0",
		    "tag_name": "v1.122.0",
		    "description": "the notes",
		    "released_at": "2026-10-09T06:40:31.189Z",
		    "upcoming_release": false,
		    "assets": {"count": 2, "links": [{"name": "glab_1.122.0_linux_amd64.tar.gz"}]},
		    "_links": {"self": "https://gitlab.com/gitlab-org/cli/-/releases/v1.122.0"}
		  },
		  {
		    "name": "v1.123.0",
		    "tag_name": "v1.123.0",
		    "upcoming_release": true
		  }
		]`))
	}))
	defer server.Close()

	client := New(toServer(server))
	releases, err := client.ListReleases(context.Background(), "gitlab.com", "gitlab-org/cli", 2, 30)
	if err != nil {
		t.Fatal(err)
	}
	// A project is one escaped path segment: GitLab reads <namespace>/<project> as a
	// single id.
	if got.EscapedPath() != "/api/v4/projects/gitlab-org%2Fcli/releases" {
		t.Fatalf("wanted the project's releases, got %s", got.EscapedPath())
	}
	// per_page rather than limit: this is GitLab's API rather than Gitea's.
	if q := got.Query(); q.Get("page") != "2" || q.Get("per_page") != "30" {
		t.Fatalf("wanted page 2 of 30, got %s", got.RawQuery)
	}
	if len(releases) != 2 {
		t.Fatalf("wanted both releases, got %d", len(releases))
	}
	first := releases[0]
	if first.TagName != "v1.122.0" || first.Name != "v1.122.0" {
		t.Fatalf("wanted the release the instance answered with, got %+v", first)
	}
	// description and _links.self are what the preview is written from.
	if first.Body != "the notes" || first.HTMLURL != "https://gitlab.com/gitlab-org/cli/-/releases/v1.122.0" {
		t.Fatalf("wanted what the preview is written from, got %+v", first)
	}
	if first.Draft || first.Prerelease {
		t.Fatalf("wanted a release to install from, got %+v", first)
	}
	// upcoming_release is GitLab's answer to the question draft answers elsewhere: a
	// release dated in the future is not one to install from.
	if !releases[1].Draft {
		t.Fatalf("wanted an upcoming release to be left out, got %+v", releases[1])
	}
}

// A project in subgroups is named by its namespace, which is still one segment to GitLab.
func TestClient_ListReleases_subgroups(t *testing.T) {
	t.Parallel()
	var got *url.URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := New(toServer(server))
	if _, err := client.ListReleases(context.Background(), "gitlab.com", "gitlab-org/security/cli", 1, 100); err != nil {
		t.Fatal(err)
	}
	if got.EscapedPath() != "/api/v4/projects/gitlab-org%2Fsecurity%2Fcli/releases" {
		t.Fatalf("wanted the namespace escaped into one segment, got %s", got.EscapedPath())
	}
}

// What the instance said about a refusal is in the error: GitLab says which of a missing
// project and a private one it was.
func TestClient_ListReleases_refused(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"404 Project Not Found"}`))
	}))
	defer server.Close()

	client := New(toServer(server))
	_, err := client.ListReleases(context.Background(), "gitlab.com", "an-owner/a-project", 1, 100)
	if err == nil {
		t.Fatal("an error must be returned")
	}
	if !strings.Contains(err.Error(), "404 Project Not Found") {
		t.Fatalf("wanted the error to carry what the instance said, got %v", err)
	}
}

func TestClient_MaxPerPage(t *testing.T) {
	t.Parallel()
	if n := New(http.DefaultClient).MaxPerPage(); n != MaxPerPage {
		t.Fatalf("wanted %d, got %d", MaxPerPage, n)
	}
}

// toServer is an HTTP client that answers from the test server whatever instance it is
// asked for, so that the request the client builds is the one under test.
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
