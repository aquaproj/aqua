package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestSLSAProvenance_GetEnabled(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		slsa     *registry.SLSAProvenance
		expected bool
	}{
		{
			name:     "nil slsa",
			slsa:     nil,
			expected: false,
		},
		{
			name:     "empty slsa config",
			slsa:     &registry.SLSAProvenance{},
			expected: false,
		},
		{
			name: "enabled explicitly true",
			slsa: &registry.SLSAProvenance{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled explicitly false",
			slsa: &registry.SLSAProvenance{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "enabled by having type",
			slsa: &registry.SLSAProvenance{
				Type: "github_release",
			},
			expected: true,
		},
		{
			name: "disabled even with type when explicitly false",
			slsa: &registry.SLSAProvenance{
				Enabled: boolPtr(false),
				Type:    "github_release",
			},
			expected: false,
		},
		{
			name: "full slsa configuration",
			slsa: &registry.SLSAProvenance{
				Enabled:   boolPtr(true),
				Type:      "github_release",
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     stringPtr("{{.Asset}}.intoto.jsonl"),
				SourceURI: stringPtr("github.com/owner/repo"),
				SourceTag: "{{.Version}}",
			},
			expected: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.slsa.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestSLSAProvenance_GetSourceURI(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		slsa     *registry.SLSAProvenance
		expected string
	}{
		{
			name: "explicit source URI",
			slsa: &registry.SLSAProvenance{
				SourceURI: stringPtr("github.com/custom/repo"),
			},
			expected: "github.com/custom/repo",
		},
		{
			name: "derived from repo owner and name",
			slsa: &registry.SLSAProvenance{
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
			},
			expected: "github.com/aquaproj/aqua",
		},
		{
			name: "explicit source URI takes precedence",
			slsa: &registry.SLSAProvenance{
				SourceURI: stringPtr("github.com/explicit/uri"),
				RepoOwner: "ignored",
				RepoName:  "ignored",
			},
			expected: "github.com/explicit/uri",
		},
		{
			name: "empty repo info",
			slsa: &registry.SLSAProvenance{
				RepoOwner: "",
				RepoName:  "",
			},
			expected: "github.com//",
		},
		{
			name: "partial repo info",
			slsa: &registry.SLSAProvenance{
				RepoOwner: "owner",
				RepoName:  "",
			},
			expected: "github.com/owner/",
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.slsa.GetSourceURI()
			if result != d.expected {
				t.Errorf("expected %q, got %q", d.expected, result)
			}
		})
	}
}

func TestSLSAProvenance_ToDownloadedFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		slsa     *registry.SLSAProvenance
		expected *registry.DownloadedFile
	}{
		{
			name: "github_release slsa",
			slsa: &registry.SLSAProvenance{
				Type:      "github_release",
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     stringPtr("{{.Asset}}.intoto.jsonl"),
			},
			expected: &registry.DownloadedFile{
				Type:      "github_release",
				RepoOwner: "owner",
				RepoName:  "repo",
				Asset:     stringPtr("{{.Asset}}.intoto.jsonl"),
			},
		},
		{
			name: "http slsa",
			slsa: &registry.SLSAProvenance{
				Type: "http",
				URL:  stringPtr("https://example.com/provenance.jsonl"),
			},
			expected: &registry.DownloadedFile{
				Type: "http",
				URL:  stringPtr("https://example.com/provenance.jsonl"),
			},
		},
		{
			name: "full slsa config",
			slsa: &registry.SLSAProvenance{
				Enabled:   boolPtr(true),
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     stringPtr("aqua_{{.OS}}_{{.Arch}}.tar.gz.intoto.jsonl"),
				SourceURI: stringPtr("github.com/aquaproj/aqua"),
				SourceTag: "{{.Version}}",
			},
			expected: &registry.DownloadedFile{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     stringPtr("aqua_{{.OS}}_{{.Arch}}.tar.gz.intoto.jsonl"),
			},
		},
		{
			name:     "empty slsa",
			slsa:     &registry.SLSAProvenance{},
			expected: &registry.DownloadedFile{},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.slsa.ToDownloadedFile()
			if diff := cmp.Diff(d.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSLSAProvenance_GetDownloadedFile(t *testing.T) {
	t.Parallel()
	// This is an alias for ToDownloadedFile, so we just test it works
	slsa := &registry.SLSAProvenance{
		Type:      "github_release",
		RepoOwner: "owner",
		RepoName:  "repo",
		Asset:     stringPtr("provenance.jsonl"),
	}

	result1 := slsa.ToDownloadedFile()
	result2 := slsa.GetDownloadedFile()

	if diff := cmp.Diff(result1, result2); diff != "" {
		t.Errorf("GetDownloadedFile should return same as ToDownloadedFile (-want +got):\n%s", diff)
	}
}
