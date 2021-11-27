package controller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_mergedRegistry_GetRegistry(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		registry *mergedRegistry
		isErr    bool
		exp      Registry
	}{
		{
			title: "local",
			registry: &mergedRegistry{
				Type: registryTypeLocal,
				Name: "foo",
				Path: "foo.yaml",
			},
			exp: &LocalRegistry{
				Name: "foo",
				Path: "foo.yaml",
			},
		},
		{
			title: "github_content",
			registry: &mergedRegistry{
				Type:      registryTypeGitHubContent,
				Name:      "foo",
				RepoOwner: "aquaproj",
				RepoName:  "ci-info",
				Ref:       "v1.0.0",
				Path:      "registry.yaml",
			},
			exp: &GitHubContentRegistry{
				Name:      "foo",
				RepoOwner: "aquaproj",
				RepoName:  "ci-info",
				Ref:       "v1.0.0",
				Path:      "registry.yaml",
			},
		},
		{
			title: "unknown",
			registry: &mergedRegistry{
				Type: "unknown",
				Name: "foo",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			registry, err := d.registry.GetRegistry()
			if d.isErr {
				if err == nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if diff := cmp.Diff(d.exp, registry); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
