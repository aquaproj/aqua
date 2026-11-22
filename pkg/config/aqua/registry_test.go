package aqua_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func TestRegistry_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		registry *aqua.Registry
		isErr    bool
	}{
		{
			title: pkgTypeGitHubContent,
			registry: &aqua.Registry{
				RepoOwner: regOwnerAquaproj,
				RepoName:  regNameAquaRegistry,
				Ref:       versionV080,
				Path:      pkgFooYaml,
				Type:      pkgTypeGitHubContent,
			},
		},
		{
			title: "github_content repo_owner is required",
			registry: &aqua.Registry{
				RepoName: regNameAquaRegistry,
				Ref:      versionV080,
				Path:     pkgFooYaml,
				Type:     pkgTypeGitHubContent,
			},
			isErr: true,
		},
		{
			title: "github_content repo_name is required",
			registry: &aqua.Registry{
				RepoOwner: regOwnerAquaproj,
				Ref:       versionV080,
				Path:      pkgFooYaml,
				Type:      pkgTypeGitHubContent,
			},
			isErr: true,
		},
		{
			title: "github_content ref is required",
			registry: &aqua.Registry{
				RepoOwner: regOwnerAquaproj,
				RepoName:  regNameAquaRegistry,
				Path:      pkgFooYaml,
				Type:      pkgTypeGitHubContent,
			},
			isErr: true,
		},
		{
			title: regTypeLocal,
			registry: &aqua.Registry{
				Path: pkgFooYaml,
				Type: regTypeLocal,
			},
		},
		{
			title: "local path is required",
			registry: &aqua.Registry{
				Type: regTypeLocal,
			},
			isErr: true,
		},
		{
			title: "invalid type",
			registry: &aqua.Registry{
				Type: "invalid-type",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.registry.Validate(); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestRegistry_FilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *aqua.Registry
		rootDir     string
		homeDir     string
		cfgFilePath string
		isErr       bool
	}{
		{
			title:       "normal",
			exp:         "ci/foo.yaml",
			rootDir:     "/root/.aqua",
			homeDir:     "/root",
			cfgFilePath: "ci/aqua.yaml",
			registry: &aqua.Registry{
				Path: pkgFooYaml,
				Type: regTypeLocal,
			},
		},
		{
			title:   pkgTypeGitHubContent,
			exp:     "/root/.aqua/registries/github_content/github.com/aquaproj/aqua-registry/v0.8.0/foo.yaml",
			rootDir: "/root/.aqua",
			registry: &aqua.Registry{
				RepoOwner: regOwnerAquaproj,
				RepoName:  regNameAquaRegistry,
				Ref:       versionV080,
				Path:      pkgFooYaml,
				Type:      pkgTypeGitHubContent,
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			p, err := d.registry.FilePath(d.rootDir, d.cfgFilePath)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}

func TestRegistry_CommitFilePath(t *testing.T) { //nolint:funlen
	t.Parallel()
	commit := strings.Repeat("ab", 20)
	rgst := &aqua.Registry{
		Type:      pkgTypeGitHubContent,
		RepoOwner: "corp",
		RepoName:  "registry",
		Ref:       "v1.0.0",
		Path:      "dir/registry.yaml",
	}
	data := []struct {
		title    string
		registry *aqua.Registry
		commit   string
		exp      string
		isErr    bool
	}{
		{
			title:    "SHA-1",
			registry: rgst,
			commit:   commit,
			exp:      "/root/.aqua/registries/github_content_commit/github.com/corp/registry/" + commit + "/dir/registry.yaml",
		},
		{
			title:    "an uppercase commit names the same path",
			registry: rgst,
			commit:   strings.ToUpper(commit),
			exp:      "/root/.aqua/registries/github_content_commit/github.com/corp/registry/" + commit + "/dir/registry.yaml",
		},
		{
			title:    "SHA-256",
			registry: rgst,
			commit:   strings.Repeat("0f", 32),
			exp:      "/root/.aqua/registries/github_content_commit/github.com/corp/registry/" + strings.Repeat("0f", 32) + "/dir/registry.yaml",
		},
		{
			title:    "a commit that would leave the directory",
			registry: rgst,
			commit:   "../../../../github_content/github.com/corp/registry/v1.0",
			isErr:    true,
		},
		{
			title:    "an abbreviated commit",
			registry: rgst,
			commit:   commit[:7],
			isErr:    true,
		},
		{
			title:    "no commit",
			registry: rgst,
			isErr:    true,
		},
		{
			title:    "a local registry",
			registry: &aqua.Registry{Type: regTypeLocal, Path: "registry.yaml"},
			commit:   commit,
			isErr:    true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			p, err := d.registry.CommitFilePath("/root/.aqua", d.commit)
			if d.isErr {
				if err == nil {
					t.Fatalf("an error must be returned, got %s", p)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if p != filepath.FromSlash(d.exp) {
				t.Fatalf("got %s, want %s", p, d.exp)
			}
		})
	}
}
