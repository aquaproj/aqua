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
			name: caseEnabledTrue,
			minisign: &registry.Minisign{
				Enabled: new(true),
			},
			expected: true,
		},
		{
			name: caseEnabledFalse,
			minisign: &registry.Minisign{
				Enabled: new(false),
			},
			expected: false,
		},
		{
			name: "full minisign configuration",
			minisign: &registry.Minisign{
				Enabled:   new(true),
				Type:      pkgTypeGitHubRelease,
				Asset:     new("{{.Asset}}.minisig"),
				PublicKey: "RWS...",
			},
			expected: true,
		},
		{
			name: "disabled minisign configuration",
			minisign: &registry.Minisign{
				Enabled:   new(false),
				Type:      pkgTypeHTTP,
				URL:       new("https://example.com/signature.minisig"),
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
				Type:      pkgTypeGitHubRelease,
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     new("{{.Asset}}.minisig"),
			},
			expected: &registry.DownloadedFile{
				Type:      pkgTypeGitHubRelease,
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     new("{{.Asset}}.minisig"),
			},
		},
		{
			name: "http minisign",
			minisign: &registry.Minisign{
				Type: pkgTypeHTTP,
				URL:  new("https://example.com/signature.minisig"),
			},
			expected: &registry.DownloadedFile{
				Type: pkgTypeHTTP,
				URL:  new("https://example.com/signature.minisig"),
			},
		},
		{
			name: "full minisign config",
			minisign: &registry.Minisign{
				Enabled:   new(true),
				Type:      pkgTypeGitHubRelease,
				RepoOwner: repoOwnerAquaproj,
				RepoName:  pkgNameAqua,
				Asset:     new("aqua_{{.OS}}_{{.Arch}}.tar.gz.minisig"),
				PublicKey: "RWSomeBase64EncodedKey...",
			},
			expected: &registry.DownloadedFile{
				Type:      pkgTypeGitHubRelease,
				RepoOwner: repoOwnerAquaproj,
				RepoName:  pkgNameAqua,
				Asset:     new("aqua_{{.OS}}_{{.Arch}}.tar.gz.minisig"),
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
