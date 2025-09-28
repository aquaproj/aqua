//nolint:funlen
package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
)

func TestGitHubArtifactAttestations_GetEnabled(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		ghaa     *registry.GitHubArtifactAttestations
		expected bool
	}{
		{
			name:     "nil github artifact attestations",
			ghaa:     nil,
			expected: false,
		},
		{
			name:     "empty config (default enabled)",
			ghaa:     &registry.GitHubArtifactAttestations{},
			expected: true,
		},
		{
			name: "enabled explicitly true",
			ghaa: &registry.GitHubArtifactAttestations{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled explicitly false",
			ghaa: &registry.GitHubArtifactAttestations{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "full configuration enabled",
			ghaa: &registry.GitHubArtifactAttestations{
				Enabled:         boolPtr(true),
				PredicateType:   "https://slsa.dev/provenance/v0.2",
				SignerWorkflow2: ".github/workflows/release.yml@refs/tags/{{.Version}}",
			},
			expected: true,
		},
		{
			name: "full configuration disabled",
			ghaa: &registry.GitHubArtifactAttestations{
				Enabled:         boolPtr(false),
				PredicateType:   "https://slsa.dev/provenance/v0.2",
				SignerWorkflow2: ".github/workflows/release.yml@refs/tags/{{.Version}}",
			},
			expected: false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.ghaa.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestGitHubArtifactAttestations_SignerWorkflow(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		ghaa     *registry.GitHubArtifactAttestations
		expected string
	}{
		{
			name:     "nil github artifact attestations",
			ghaa:     nil,
			expected: "",
		},
		{
			name:     "empty config",
			ghaa:     &registry.GitHubArtifactAttestations{},
			expected: "",
		},
		{
			name: "only new signer workflow field",
			ghaa: &registry.GitHubArtifactAttestations{
				SignerWorkflow2: ".github/workflows/release.yml@refs/tags/v1.0.0",
			},
			expected: ".github/workflows/release.yml@refs/tags/v1.0.0",
		},
		{
			name: "only deprecated signer workflow field",
			ghaa: &registry.GitHubArtifactAttestations{
				SignerWorkflow3: ".github/workflows/deprecated.yml@refs/tags/v1.0.0",
			},
			expected: ".github/workflows/deprecated.yml@refs/tags/v1.0.0",
		},
		{
			name: "both fields - new takes precedence",
			ghaa: &registry.GitHubArtifactAttestations{
				SignerWorkflow2: ".github/workflows/new.yml@refs/tags/v1.0.0",
				SignerWorkflow3: ".github/workflows/deprecated.yml@refs/tags/v1.0.0",
			},
			expected: ".github/workflows/new.yml@refs/tags/v1.0.0",
		},
		{
			name: "empty new field, falls back to deprecated",
			ghaa: &registry.GitHubArtifactAttestations{
				SignerWorkflow2: "",
				SignerWorkflow3: ".github/workflows/fallback.yml@refs/tags/v1.0.0",
			},
			expected: ".github/workflows/fallback.yml@refs/tags/v1.0.0",
		},
		{
			name: "full configuration with template",
			ghaa: &registry.GitHubArtifactAttestations{
				Enabled:         boolPtr(true),
				PredicateType:   "https://slsa.dev/provenance/v0.2",
				SignerWorkflow2: ".github/workflows/release.yml@refs/tags/{{.Version}}",
			},
			expected: ".github/workflows/release.yml@refs/tags/{{.Version}}",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.ghaa.SignerWorkflow()
			if result != d.expected {
				t.Errorf("expected %q, got %q", d.expected, result)
			}
		})
	}
}
