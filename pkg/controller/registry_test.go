package controller_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/controller"
)

func TestRegistry_GetType(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		exp      string
		registry *controller.Registry
	}{
		{
			title: "github_content",
			exp:   "github_content",
			registry: &controller.Registry{
				Type: "github_content",
			},
		},
		{
			title: "local",
			exp:   "local",
			registry: &controller.Registry{
				Type: "local",
			},
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

func TestRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *controller.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:   "github_content",
			exp:     "/root/.aqua/registries/github_content/github.com/aquaproj/aqua-registry/v0.8.0/foo.yaml",
			rootDir: "/root/.aqua",
			registry: &controller.Registry{
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
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

func TestRegistry_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *controller.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title: "local",
			exp:   "foo",
			registry: &controller.Registry{
				Type: "local",
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

func TestLocalRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *controller.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:       "normal",
			exp:         "ci/foo.yaml",
			rootDir:     "/root/.aqua",
			cfgFilePath: "ci/aqua.yaml",
			registry: &controller.Registry{
				Path: "foo.yaml",
				Type: "local",
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
