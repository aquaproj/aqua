package aqua_test

import (
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
