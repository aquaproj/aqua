//nolint:funlen
package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func TestChecksum_GetReplacements(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		checksum *registry.Checksum
		expected registry.Replacements
	}{
		{
			name:     "nil checksum",
			checksum: nil,
			expected: nil,
		},
		{
			name:     "empty checksum",
			checksum: &registry.Checksum{},
			expected: nil,
		},
		{
			name: "checksum with replacements",
			checksum: &registry.Checksum{
				Replacements: registry.Replacements{
					"linux":  "Linux",
					"darwin": "macOS",
				},
			},
			expected: registry.Replacements{
				"linux":  "Linux",
				"darwin": "macOS",
			},
		},
		{
			name: "checksum with empty replacements map",
			checksum: &registry.Checksum{
				Replacements: registry.Replacements{},
			},
			expected: registry.Replacements{},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetReplacements()
			if d.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(d.expected) {
				t.Errorf("expected length %d, got %d", len(d.expected), len(result))
				return
			}
			for k, v := range d.expected {
				if result[k] != v {
					t.Errorf("expected %q for key %q, got %q", v, k, result[k])
				}
			}
		})
	}
}

func TestChecksum_GetEnabled(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		checksum *registry.Checksum
		expected bool
	}{
		{
			name:     "nil checksum",
			checksum: nil,
			expected: false,
		},
		{
			name:     "enabled is nil (default true)",
			checksum: &registry.Checksum{},
			expected: true,
		},
		{
			name: "enabled is true",
			checksum: &registry.Checksum{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled is false",
			checksum: &registry.Checksum{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "full checksum configuration with enabled true",
			checksum: &registry.Checksum{
				Type:      "github_release",
				Asset:     "checksums.txt",
				Algorithm: "sha256",
				Enabled:   boolPtr(true),
				Replacements: registry.Replacements{
					"linux": "Linux",
				},
			},
			expected: true,
		},
		{
			name: "full checksum configuration with enabled false",
			checksum: &registry.Checksum{
				Type:      "http",
				URL:       "https://example.com/checksums.txt",
				Algorithm: "sha512",
				Enabled:   boolPtr(false),
			},
			expected: false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestChecksum_GetAlgorithm(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		checksum  *registry.Checksum
		expected  string
		expectErr bool
	}{
		{
			name:     "nil checksum",
			checksum: nil,
			expected: "sha256", // Default algorithm
		},
		{
			name:     "empty algorithm (default)",
			checksum: &registry.Checksum{},
			expected: "",
		},
		{
			name: "explicit sha256",
			checksum: &registry.Checksum{
				Algorithm: "sha256",
			},
			expected: "sha256",
		},
		{
			name: "sha512 algorithm",
			checksum: &registry.Checksum{
				Algorithm: "sha512",
			},
			expected: "sha512",
		},
		{
			name: "md5 algorithm",
			checksum: &registry.Checksum{
				Algorithm: "md5",
			},
			expected: "md5",
		},
		{
			name: "sha1 algorithm",
			checksum: &registry.Checksum{
				Algorithm: "sha1",
			},
			expected: "sha1",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetAlgorithm()
			if result != d.expected {
				t.Errorf("expected %q, got %q", d.expected, result)
			}
		})
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

func TestChecksum_GetCosign(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name     string
		checksum *registry.Checksum
		expected *registry.Cosign
	}{
		{
			name:     "nil checksum",
			checksum: nil,
			expected: nil,
		},
		{
			name:     "checksum without cosign",
			checksum: &registry.Checksum{},
			expected: nil,
		},
		{
			name: "checksum with cosign",
			checksum: &registry.Checksum{
				Cosign: &registry.Cosign{
					Enabled: boolPtr(true),
				},
			},
			expected: &registry.Cosign{
				Enabled: boolPtr(true),
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetCosign()
			if d.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Error("expected non-nil cosign")
				return
			}
			if result.GetEnabled() != d.expected.GetEnabled() {
				t.Errorf("expected enabled %v, got %v", d.expected.GetEnabled(), result.GetEnabled())
			}
		})
	}
}

func TestChecksum_GetMinisign(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name     string
		checksum *registry.Checksum
		expected *registry.Minisign
	}{
		{
			name:     "nil checksum",
			checksum: nil,
			expected: nil,
		},
		{
			name:     "checksum without minisign",
			checksum: &registry.Checksum{},
			expected: nil,
		},
		{
			name: "checksum with minisign",
			checksum: &registry.Checksum{
				Minisign: &registry.Minisign{
					Enabled: boolPtr(true),
				},
			},
			expected: &registry.Minisign{
				Enabled: boolPtr(true),
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetMinisign()
			if d.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Error("expected non-nil minisign")
				return
			}
			if result.GetEnabled() != d.expected.GetEnabled() {
				t.Errorf("expected enabled %v, got %v", d.expected.GetEnabled(), result.GetEnabled())
			}
		})
	}
}

func TestChecksum_GetGitHubArtifactAttestations(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name     string
		checksum *registry.Checksum
		expected *registry.GitHubArtifactAttestations
	}{
		{
			name:     "nil checksum",
			checksum: nil,
			expected: nil,
		},
		{
			name:     "checksum without github artifact attestations",
			checksum: &registry.Checksum{},
			expected: nil,
		},
		{
			name: "checksum with github artifact attestations",
			checksum: &registry.Checksum{
				GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{
					Enabled: boolPtr(true),
				},
			},
			expected: &registry.GitHubArtifactAttestations{
				Enabled: boolPtr(true),
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.checksum.GetGitHubArtifactAttestations()
			if d.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Error("expected non-nil github artifact attestations")
				return
			}
			if result.GetEnabled() != d.expected.GetEnabled() {
				t.Errorf("expected enabled %v, got %v", d.expected.GetEnabled(), result.GetEnabled())
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
