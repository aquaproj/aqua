package config_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"gopkg.in/yaml.v2"
)

func TestRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *config.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:   "github_content",
			exp:     "/root/.aqua/registries/github_content/github.com/aquaproj/aqua-registry/v0.8.0/foo.yaml",
			rootDir: "/root/.aqua",
			registry: &config.Registry{
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

func TestLocalRegistry_GetFilePath(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		exp         string
		registry    *config.Registry
		rootDir     string
		cfgFilePath string
	}{
		{
			title:       "normal",
			exp:         "ci/foo.yaml",
			rootDir:     "/root/.aqua",
			cfgFilePath: "ci/aqua.yaml",
			registry: &config.Registry{
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

func downloadTestFile(uri, tempDir string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("create a request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	fileName := "registry.yaml"
	if filepath.Ext(uri) == ".json" {
		fileName = "registry.json"
	}
	filePath := filepath.Join(tempDir, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create a file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write a response body to a file: %w", err)
	}
	return filePath, nil
}

func BenchmarkReadRegistry(b *testing.B) {
	b.Run("yaml", func(b *testing.B) {
		registryYAML, err := downloadTestFile("https://raw.githubusercontent.com/aquaproj/aqua-registry/v2.11.1/registry.yaml", b.TempDir())
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			func() {
				f, err := os.Open(registryYAML)
				if err != nil {
					b.Fatal(err)
				}
				defer f.Close()
				registry := &config.RegistryContent{}
				if err := yaml.NewDecoder(f).Decode(registry); err != nil {
					b.Fatal(err)
				}
			}()
		}
	})
	b.Run("json", func(b *testing.B) {
		registryJSON, err := downloadTestFile("https://raw.githubusercontent.com/aquaproj/aqua-registry/v2.11.1/registry.json", b.TempDir())
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			func() {
				f, err := os.Open(registryJSON)
				if err != nil {
					b.Fatal(err)
				}
				defer f.Close()
				registry := &config.RegistryContent{}
				if err := json.NewDecoder(f).Decode(registry); err != nil {
					b.Fatal(err)
				}
			}()
		}
	})
}

func TestRegistry_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title    string
		registry *config.Registry
		isErr    bool
	}{
		{
			title: "github_content",
			registry: &config.Registry{
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
		},
		{
			title: "github_content repo_owner is required",
			registry: &config.Registry{
				RepoName: "aqua-registry",
				Ref:      "v0.8.0",
				Path:     "foo.yaml",
				Type:     "github_content",
			},
			isErr: true,
		},
		{
			title: "github_content repo_name is required",
			registry: &config.Registry{
				RepoOwner: "aquaproj",
				Ref:       "v0.8.0",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
			isErr: true,
		},
		{
			title: "github_content ref is required",
			registry: &config.Registry{
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Path:      "foo.yaml",
				Type:      "github_content",
			},
			isErr: true,
		},
		{
			title: "local",
			registry: &config.Registry{
				Path: "foo.yaml",
				Type: "local",
			},
		},
		{
			title: "local path is required",
			registry: &config.Registry{
				Type: "local",
			},
			isErr: true,
		},
		{
			title: "invalid type",
			registry: &config.Registry{
				Type: "invalid-type",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
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
