package g2_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
)

const templateID = "1795000001"

// templateDefinition is the registry.yaml of a package served as a template: no list of
// versions anywhere, one asset name until v2, another since.
const templateDefinition = `
name: example/cli
type: http
url: https://example.com/{{.Version}}/cli_{{.OS}}_{{.Arch}}.tar.gz
files:
  - name: cli
repo_id: 1234
all_assets_filter: "true"
template:
  reason: the vendor publishes no list of versions
  test_versions:
    - v2.0.0
    - v1.0.0
version_overrides:
  - version_constraint: semver(">= 2.0.0")
    url: https://example.com/{{.Version}}/cli-{{.OS}}-{{.Arch}}.tar.gz
  - version_constraint: semver(">= 1.0.0")
`

func readDefinition(t *testing.T, s string) *g2.Config {
	t.Helper()
	cfg := &g2.Config{}
	if err := yaml.Unmarshal([]byte(s), cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

// templateJSON is what ar2 writes for templateDefinition.
func templateJSON(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(g2.NewTemplate(readDefinition(t, templateDefinition)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// templateDownloader serves a registry the way raw.githubusercontent.com does for a public
// repository: a file it doesn't have is ErrGitHubContentNotFound. offline makes every request
// fail some other way.
type templateDownloader struct {
	files   map[string]string
	offline bool
	calls   map[string]int
}

func (d *templateDownloader) DownloadGitHubContentFile(_ context.Context, _ *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	if d.calls == nil {
		d.calls = map[string]int{}
	}
	d.calls[param.Path]++
	if d.offline {
		return nil, errors.New("dial tcp: lookup raw.githubusercontent.com: no such host")
	}
	content, ok := d.files[param.Path]
	if !ok {
		if !param.Public {
			return nil, errors.New("the API was asked")
		}
		return nil, domain.ErrGitHubContentNotFound
	}
	return &domain.GitHubContentFile{ReadCloser: io.NopCloser(strings.NewReader(content))}, nil
}

func newTemplateDownloader(t *testing.T) *templateDownloader {
	t.Helper()
	return &templateDownloader{files: map[string]string{
		g2.NamesFileName: `{"ids":{"example/cli":"` + templateID + `"}}`,
		g2.PackageDir(templateID) + "/" + g2.TemplateFileName: templateJSON(t),
	}}
}

// What ar2 writes leaves out what only generating versions reads, and how the package is
// checked, while keeping everything that installs it.
func TestNewTemplate(t *testing.T) {
	t.Parallel()
	got := map[string]any{}
	if err := json.Unmarshal([]byte(templateJSON(t)), &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"repo_id", "RepoID", "all_assets_filter", "AllAssetsFilter", "template", "Template"} {
		if _, ok := got[key]; ok {
			t.Errorf("the template has %s", key)
		}
	}
	for _, key := range []string{"name", "type", "url", "files", "version_overrides"} {
		if _, ok := got[key]; !ok {
			t.Errorf("the template lacks %s", key)
		}
	}
}

// Versioned answers what SetVersion does, and what it answers survives aqua-registry's own
// resolution, which reads version_overrides the other way round and is applied again by
// whatever resolves the result.
func TestConfig_Versioned(t *testing.T) {
	t.Parallel()
	cfg := readDefinition(t, `
name: example/cli
type: http
url: https://example.com/top/{{.Version}}
version_prefix: cli/
version_overrides:
  - version_constraint: semver(">= 2.0.0")
    url: https://example.com/new/{{.Version}}
  - version_constraint: semver(">= 1.0.0")
    version_prefix: ""
    url: https://example.com/unprefixed/{{.Version}}
`)
	logger := discardLogger()
	for _, v := range []string{"cli/v2.1.0", "cli/v1.5.0", "v1.5.0", "cli/v0.1.0", "latest"} {
		want, err := cfg.SetVersion(logger, v)
		if err != nil {
			t.Fatal(err)
		}
		versioned, err := cfg.Versioned(logger, v)
		if err != nil {
			t.Fatal(err)
		}
		again, err := versioned.SetVersion(logger, v)
		if err != nil {
			t.Fatal(err)
		}
		if versioned.URL != want.URL || again.URL != want.URL {
			t.Errorf("%s: got %v then %v, want %v", v, versioned.URL, again.URL, want.URL)
		}
		if again.VersionPrefix != want.VersionPrefix {
			t.Errorf("%s: version_prefix is %q, want %q", v, again.VersionPrefix, want.VersionPrefix)
		}
	}
}

// A version with no registry.json is expanded from the package's template, for the
// environments asked for, and comes back with the definition its checksums are fetched from.
func TestClient_Resolve_template(t *testing.T) {
	t.Parallel()
	dl := newTemplateDownloader(t)
	resolved, err := g2.New(dl, nil, "", "").Resolve(t.Context(), discardLogger(),
		"example/cli", "v2.1.0", []string{"linux/amd64"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.PkgInfo == nil {
		t.Fatal("a template's definition must be returned for its checksums")
	}
	if len(resolved.Packages) != 1 {
		t.Fatalf("got %d entries, want the one environment asked for", len(resolved.Packages))
	}
	pkg := resolved.Packages[0]
	if diff := cmp.Diff(&lockfile.Registry{Type: "github_content", RepoOwner: "aquaproj", RepoName: "aqua-registry-g2"}, pkg.Registry); diff != "" {
		t.Errorf("the registry is wrong (-want +got):\n%s", diff)
	}
	if want := "https://example.com/v2.1.0/cli-linux-amd64.tar.gz"; pkg.URL != want {
		t.Errorf("url is %q, want %q (the newest override)", pkg.URL, want)
	}
	if pkg.Checksum != "" {
		t.Errorf("a template carries no checksum, got %q", pkg.Checksum)
	}
}

// A package with neither the version's registry.json nor a template fails, and says so.
func TestClient_Resolve_noTemplate(t *testing.T) {
	t.Parallel()
	dl := newTemplateDownloader(t)
	delete(dl.files, g2.PackageDir(templateID)+"/"+g2.TemplateFileName)
	_, err := g2.New(dl, nil, "", "").Resolve(t.Context(), discardLogger(), "example/cli", "v2.1.0", nil)
	if !errors.Is(err, g2.ErrNoTemplate) {
		t.Fatalf("want ErrNoTemplate, got %v", err)
	}
}

// Install and exec take a template cached within a day without asking the registry, and one
// of any age when the registry can't be read -- which also waits another day before asking
// again. The lock update command reads it fresh and fails rather than locking an old copy.
func TestClient_Definitions_cache(t *testing.T) {
	t.Parallel()
	cache := g2.NewCache(&config.Param{CacheDir: t.TempDir()})
	logger := discardLogger()
	if _, err := g2.New(newTemplateDownloader(t), cache, "", "").Resolve(t.Context(), logger,
		"example/cli", "v2.1.0", nil); err != nil {
		t.Fatal(err)
	}

	offline := newTemplateDownloader(t)
	offline.offline = true
	// Cached the names table as well, so only the files themselves are asked for.
	client := g2.New(offline, cache, "", "")

	if _, err := client.Definitions(t.Context(), logger, "example/cli", "v2.1.0", nil); err != nil {
		t.Fatalf("a fresh cached template answers: %v", err)
	}
	if len(offline.calls) != 0 {
		t.Errorf("asked the registry %v, want nothing", offline.calls)
	}

	path := cache.TemplatePath("aquaproj", "aqua-registry-g2", "example/cli")
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Definitions(t.Context(), logger, "example/cli", "v2.1.0", nil); err != nil {
		t.Fatalf("an old cached template answers offline: %v", err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(stat.ModTime()) > time.Hour {
		t.Error("the age of the cached template must be reset, so the next run doesn't ask again")
	}

	if _, err := client.Resolve(t.Context(), logger, "example/cli", "v2.1.0", nil); err == nil {
		t.Error("the lock update command must not lock a cached template it couldn't refresh")
	}
}

// The template is what registry.PackageInfo reads, so whatever aqua resolves is what the
// definition said.
func TestParseTemplate(t *testing.T) {
	t.Parallel()
	tmpl, err := g2.ParseTemplate([]byte(templateJSON(t)))
	if err != nil {
		t.Fatal(err)
	}
	want := g2.NewTemplate(readDefinition(t, templateDefinition))
	if diff := cmp.Diff(want, tmpl.PackageInfo, cmp.AllowUnexported(registry.PackageInfo{})); diff != "" {
		t.Errorf("the template doesn't round-trip (-want +got):\n%s", diff)
	}
}
