package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestMinisign_GetEnabled(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		minisign *registry.Minisign
		expected bool
	}{
		{
			name:     "nil minisign",
			minisign: nil,
			expected: false,
		},
		{
			name:     "empty minisign config (default enabled)",
			minisign: &registry.Minisign{},
			expected: true,
		},
		{
			name: "enabled explicitly true",
			minisign: &registry.Minisign{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled explicitly false",
			minisign: &registry.Minisign{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "full minisign configuration",
			minisign: &registry.Minisign{
				Enabled:   boolPtr(true),
				Type:      "github_release",
				Asset:     stringPtr("{{.Asset}}.minisig"),
				PublicKey: "RWS...",
			},
			expected: true,
		},
		{
			name: "disabled minisign configuration",
			minisign: &registry.Minisign{
				Enabled:   boolPtr(false),
				Type:      "http",
				URL:       stringPtr("https://example.com/signature.minisig"),
				PublicKey: "RWS...",
			},
			expected: false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.minisign.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestMinisign_ToDownloadedFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		minisign *registry.Minisign
		expected *registry.DownloadedFile
	}{
		{
			name: "github_release minisign",
			minisign: &registry.Minisign{
				Type:      "github_release",
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     stringPtr("{{.Asset}}.minisig"),
			},
			expected: &registry.DownloadedFile{
				Type:      "github_release",
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     stringPtr("{{.Asset}}.minisig"),
			},
		},
		{
			name: "http minisign",
			minisign: &registry.Minisign{
				Type: "http",
				URL:  stringPtr("https://example.com/signature.minisig"),
			},
			expected: &registry.DownloadedFile{
				Type: "http",
				URL:  stringPtr("https://example.com/signature.minisig"),
			},
		},
		{
			name: "full minisign config",
			minisign: &registry.Minisign{
				Enabled:   boolPtr(true),
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     stringPtr("aqua_{{.OS}}_{{.Arch}}.tar.gz.minisig"),
				PublicKey: "RWSomeBase64EncodedKey...",
			},
			expected: &registry.DownloadedFile{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     stringPtr("aqua_{{.OS}}_{{.Arch}}.tar.gz.minisig"),
			},
		},
		{
			name:     "empty minisign",
			minisign: &registry.Minisign{},
			expected: &registry.DownloadedFile{},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.minisign.ToDownloadedFile()
			if diff := cmp.Diff(d.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
