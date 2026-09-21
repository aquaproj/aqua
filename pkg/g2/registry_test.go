package g2_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/google/go-cmp/cmp"
)

// fakeDownloader answers with fixed content and records what it was asked for.
type fakeDownloader struct {
	content string
	err     error
	param   *domain.GitHubContentFileParam
}

func (d *fakeDownloader) DownloadGitHubContentFile(_ context.Context, _ *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	d.param = param
	if d.err != nil {
		return nil, d.err
	}
	return &domain.GitHubContentFile{ReadCloser: io.NopCloser(strings.NewReader(d.content))}, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestClient_Get(t *testing.T) {
	t.Parallel()
	dl := &fakeDownloader{content: `{"assets":[
		{"os":"linux","arch":"amd64","type":"github_release","repo_owner":"cli","repo_name":"cli",
		 "asset":"gh_2.1.0_linux_amd64.tar.gz","format":"tar.gz","checksum":"abc","checksum_algorithm":"sha256",
		 "files":[{"name":"gh","src":"gh_2.1.0_linux_amd64/bin/gh"}]},
		{"os":"linux","arch":"amd64","variants":{"libc":"musl"},"type":"github_release",
		 "asset":"gh_2.1.0_linux_amd64_musl.tar.gz"}]}`}

	reg, err := g2.New(dl, "", "").Get(t.Context(), discardLogger(), "cli/cli", "v2.1.0")
	if err != nil {
		t.Fatal(err)
	}

	// The package is carried by the branch and the version by the path, because
	// each package has a branch of its own.
	want := &domain.GitHubContentFileParam{
		RepoOwner: "aquaproj",
		RepoName:  "aqua-registry-g2",
		Ref:       "pkg_cli_2fcli",
		Path:      "versions/v2.1.0/registry.json",
	}
	if diff := cmp.Diff(want, dl.param); diff != "" {
		t.Errorf("the request is wrong (-want +got):\n%s", diff)
	}

	if len(reg.Assets) != 2 {
		t.Fatalf("got %d assets, want 2", len(reg.Assets))
	}
	if diff := cmp.Diff("gh_2.1.0_linux_amd64/bin/gh", reg.Assets[0].Files[0].Src); diff != "" {
		t.Errorf("files are wrong (-want +got):\n%s", diff)
	}
	// Two entries share an os and arch and are told apart by their variants.
	if diff := cmp.Diff(map[string]string{"libc": "musl"}, reg.Assets[1].Variants); diff != "" {
		t.Errorf("variants are wrong (-want +got):\n%s", diff)
	}
}

// TestClient_Get_notGenerated checks that a version g2 doesn't hold yet fails
// rather than returning nothing. The caller decides whether to fall back to another
// registry, and an empty result would look like a package supporting no environment.
func TestClient_Get_notGenerated(t *testing.T) {
	t.Parallel()
	dl := &fakeDownloader{err: errors.New("404")}
	if _, err := g2.New(dl, "", "").Get(t.Context(), discardLogger(), "cli/cli", "v9.9.9"); err == nil {
		t.Error("a version that isn't in g2 should be an error")
	}
}

func TestClient_Get_empty(t *testing.T) {
	t.Parallel()
	dl := &fakeDownloader{content: `{"assets":[]}`}
	if _, err := g2.New(dl, "", "").Get(t.Context(), discardLogger(), "cli/cli", "v2.1.0"); err == nil {
		t.Error("registry.json with no asset should be an error")
	}
}

// TestClient_Get_customRepo checks that a registry other than aqua-registry-g2 can
// be read, which is how this was tested against a scratch repository.
func TestClient_Get_customRepo(t *testing.T) {
	t.Parallel()
	dl := &fakeDownloader{content: `{"assets":[{"os":"linux","arch":"amd64","type":"github_release"}]}`}
	if _, err := g2.New(dl, "szksh-lab-2", "aqua-registry-g2-test").Get(t.Context(), discardLogger(), "cli/cli", "v2.1.0"); err != nil {
		t.Fatal(err)
	}
	if dl.param.RepoOwner != "szksh-lab-2" || dl.param.RepoName != "aqua-registry-g2-test" {
		t.Errorf("the repository is wrong: %s/%s", dl.param.RepoOwner, dl.param.RepoName)
	}
}
