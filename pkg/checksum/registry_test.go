package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

func TestRegistryID(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		registry *aqua.Registry
		expected string
	}{
		{
			name: "basic registry",
			registry: &aqua.Registry{
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry",
				Ref:       "v4.0.0",
				Path:      "",
			},
			expected: "registries/github_content/github.com/aquaproj/aqua-registry/v4.0.0",
		},
		{
			name: "registry with empty ref",
			registry: &aqua.Registry{
				RepoOwner: "example",
				RepoName:  "tools",
				Ref:       "",
				Path:      "",
			},
			expected: "registries/github_content/github.com/example/tools",
		},
		{
			name: "registry with special characters",
			registry: &aqua.Registry{
				RepoOwner: "org-name",
				RepoName:  "tool.registry",
				Ref:       "v1.2.3-beta",
				Path:      "registry.yaml",
			},
			expected: "registries/github_content/github.com/org-name/tool.registry/v1.2.3-beta/registry.yaml",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := checksum.RegistryID(d.registry)
			if result != d.expected {
				t.Errorf("Expected %s, got %s", d.expected, result)
			}
		})
	}
}

func TestCheckRegistry(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		registry *aqua.Registry
		content  []byte
		wantErr  bool
	}{
		{
			name: "new registry - no existing checksum",
			registry: &aqua.Registry{
				RepoOwner: "test",
				RepoName:  "registry",
				Ref:       "v1.0.0",
				Path:      "",
			},
			content: []byte("test content"),
			wantErr: false,
		},
		{
			name: "valid existing checksum",
			registry: &aqua.Registry{
				RepoOwner: "test2",
				RepoName:  "registry2",
				Ref:       "v1.0.0",
				Path:      "",
			},
			content: []byte("test content"),
			wantErr: false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksums := checksum.New()

			// For the second test, pre-populate with correct checksum
			if d.name == "valid existing checksum" {
				// First call to calculate and store the checksum
				err := checksum.CheckRegistry(d.registry, checksums, d.content)
				if err != nil {
					t.Fatalf("First call failed: %v", err)
				}
			}

			err := checksum.CheckRegistry(d.registry, checksums, d.content)
			if d.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !d.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCheckRegistryWithMismatch(t *testing.T) {
	t.Parallel()

	registry := &aqua.Registry{
		RepoOwner: "test",
		RepoName:  "registry",
		Ref:       "v1.0.0",
		Path:      "",
	}

	checksums := checksum.New()

	// First, establish a checksum for one content
	originalContent := []byte("original content")
	err := checksum.CheckRegistry(registry, checksums, originalContent)
	if err != nil {
		t.Fatalf("Failed to establish checksum: %v", err)
	}

	// Now try with different content - should fail
	modifiedContent := []byte("modified content")
	err = checksum.CheckRegistry(registry, checksums, modifiedContent)
	if err == nil {
		t.Error("Expected error for mismatched content but got none")
	}
}
