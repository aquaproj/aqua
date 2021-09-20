package controller_test

import (
	"testing"

	"github.com/suzuki-shunsuke/aqua/pkg/controller"
)

func TestGitHubContentRegistry_GetType(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exp      string
		registry *controller.GitHubContentRegistry
	}{
		{
			title:    "normal",
			exp:      "github_content",
			registry: &controller.GitHubContentRegistry{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if typ := d.registry.GetType(); typ != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, typ)
			}
		})
	}
}

func TestGitHubContentRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *controller.GitHubContentRegistry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:   "normal",
			exp:     "/root/.aqua/registries/github_content/github.com/suzuki-shunsuke/aqua-registry/v0.8.0/foo.yaml",
			rootDir: "/root/.aqua",
			registry: &controller.GitHubContentRegistry{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "aqua-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if p := d.registry.GetFilePath(d.rootDir, d.cfgFilePath); p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}

func TestGitHubContentRegistry_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *controller.LocalRegistry
		rootDir     string
		cfgFilePath string
	}{
		{
			title: "normal",
			exp:   "foo",
			registry: &controller.LocalRegistry{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if p := d.registry.GetName(); p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}

func TestLocalRegistry_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exp      string
		registry *controller.LocalRegistry
	}{
		{
			title: "normal",
			exp:   "foo",
			registry: &controller.LocalRegistry{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if name := d.registry.GetName(); name != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, name)
			}
		})
	}
}

func TestLocalRegistry_GetType(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exp      string
		registry *controller.LocalRegistry
	}{
		{
			title:    "normal",
			exp:      "local",
			registry: &controller.LocalRegistry{},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if typ := d.registry.GetType(); typ != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, typ)
			}
		})
	}
}

func TestLocalRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *controller.LocalRegistry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:       "normal",
			exp:         "ci/foo.yaml",
			rootDir:     "/root/.aqua",
			cfgFilePath: "ci/aqua.yaml",
			registry: &controller.LocalRegistry{
				Path: "foo.yaml",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if p := d.registry.GetFilePath(d.rootDir, d.cfgFilePath); p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}
