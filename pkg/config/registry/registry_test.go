package registry_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/yaml"
	"github.com/spf13/afero"
)

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
		yamlDecoder := yaml.NewDecoder(afero.NewOsFs())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			func() {
				registry := &registry.Config{}
				if err := yamlDecoder.ReadFile(registryYAML, registry); err != nil {
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
				registry := &registry.Config{}
				if err := json.NewDecoder(f).Decode(registry); err != nil {
					b.Fatal(err)
				}
			}()
		}
	})
}
