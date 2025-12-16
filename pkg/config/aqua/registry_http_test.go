package aqua_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func TestRegistry_ValidateHTTP(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		registry *aqua.Registry
		isErr    bool
	}{
		{
			name: "valid http registry",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/{{.Version}}/registry.yaml",
				Version: "v1.0.0",
			},
			isErr: false,
		},
		{
			name: "valid http registry with tar.gz format",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/{{.Version}}/registry.tar.gz",
				Version: "v1.0.0",
				Format:  "tar.gz",
			},
			isErr: false,
		},
		{
			name: "valid http registry with raw format",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/{{.Version}}/registry.yaml",
				Version: "v1.0.0",
				Format:  "raw",
			},
			isErr: false,
		},
		{
			name: "valid http JSON registry",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/{{.Version}}/registry.json",
				Version: "v1.0.0",
			},
			isErr: false,
		},
		{
			name: "missing URL",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				Version: "v1.0.0",
			},
			isErr: true,
		},
		{
			name: "missing version",
			registry: &aqua.Registry{
				Type: aqua.RegistryTypeHTTP,
				Name: "test",
				URL:  "https://example.com/{{.Version}}/registry.yaml",
			},
			isErr: true,
		},
		{
			name: "URL without version template",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/static/registry.yaml",
				Version: "v1.0.0",
			},
			isErr: true,
		},
		{
			name: "invalid format",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/{{.Version}}/registry.zip",
				Version: "v1.0.0",
				Format:  "zip",
			},
			isErr: true,
		},
		{
			name: "version with path traversal",
			registry: &aqua.Registry{
				Type:    aqua.RegistryTypeHTTP,
				Name:    "test",
				URL:     "https://example.com/{{.Version}}/registry.yaml",
				Version: "../etc/passwd",
			},
			isErr: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			err := d.registry.Validate()
			if d.isErr {
				if err == nil {
					t.Fatal("error must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRegistry_RenderURL(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		registry *aqua.Registry
		expected string
		isErr    bool
	}{
		{
			name: "simple version substitution",
			registry: &aqua.Registry{
				URL:     "https://example.com/{{.Version}}/registry.yaml",
				Version: "v1.0.0",
			},
			expected: "https://example.com/v1.0.0/registry.yaml",
			isErr:    false,
		},
		{
			name: "multiple version occurrences",
			registry: &aqua.Registry{
				URL:     "https://example.com/{{.Version}}/files/{{.Version}}/registry.yaml",
				Version: "v2.5.3",
			},
			expected: "https://example.com/v2.5.3/files/v2.5.3/registry.yaml",
			isErr:    false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result, err := d.registry.RenderURL()
			if d.isErr {
				if err == nil {
					t.Fatal("error must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result != d.expected {
				t.Fatalf("expected %s, got %s", d.expected, result)
			}
		})
	}
}
