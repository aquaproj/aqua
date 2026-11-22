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

func TestRegistry_LockedFilePath(t *testing.T) { //nolint:funlen
	t.Parallel()
	sha256Sum := strings.Repeat("ab", 32)
	rgst := &aqua.Registry{
		Type:      pkgTypeGitHubContent,
		RepoOwner: "corp",
		RepoName:  "registry",
		Ref:       "v1.0.0",
		Path:      "dir/registry.yaml",
	}
	data := []struct {
		title     string
		registry  *aqua.Registry
		algorithm string
		checksum  string
		exp       string
		isErr     bool
	}{
		{
			title:     "sha256",
			registry:  rgst,
			algorithm: "sha256",
			checksum:  sha256Sum,
			exp:       "/root/.aqua/registries/github_content_locked/github.com/corp/registry/v1.0.0/sha256-" + sha256Sum + "/dir/registry.yaml",
		},
		{
			title:     "an uppercase checksum names the same path",
			registry:  rgst,
			algorithm: "sha256",
			checksum:  strings.ToUpper(sha256Sum),
			exp:       "/root/.aqua/registries/github_content_locked/github.com/corp/registry/v1.0.0/sha256-" + sha256Sum + "/dir/registry.yaml",
		},
		{
			title:     "sha512",
			registry:  rgst,
			algorithm: "sha512",
			checksum:  strings.Repeat("0f", 64),
			exp:       "/root/.aqua/registries/github_content_locked/github.com/corp/registry/v1.0.0/sha512-" + strings.Repeat("0f", 64) + "/dir/registry.yaml",
		},
		{
			title:     "a checksum that would leave the directory",
			registry:  rgst,
			algorithm: "sha256",
			checksum:  "../../../../../github_content/github.com/corp/registry/v1.0.0",
			isErr:     true,
		},
		{
			title:     "a checksum of another length than the algorithm's",
			registry:  rgst,
			algorithm: "sha512",
			checksum:  sha256Sum,
			isErr:     true,
		},
		{
			title:     "a weak algorithm",
			registry:  rgst,
			algorithm: "md5",
			checksum:  strings.Repeat("ab", 16),
			isErr:     true,
		},
		{
			title:     "a local registry",
			registry:  &aqua.Registry{Type: regTypeLocal, Path: "registry.yaml"},
			algorithm: "sha256",
			checksum:  sha256Sum,
			isErr:     true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			p, err := d.registry.LockedFilePath("/root/.aqua", d.algorithm, d.checksum)
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
