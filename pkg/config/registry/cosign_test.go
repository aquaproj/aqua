//nolint:funlen
package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

func TestCosign_GetEnabled(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		cosign   *registry.Cosign
		expected bool
	}{
		{
			name:     "nil cosign",
			cosign:   nil,
			expected: false,
		},
		{
			name:     "empty cosign config",
			cosign:   &registry.Cosign{},
			expected: false,
		},
		{
			name: "enabled explicitly true",
			cosign: &registry.Cosign{
				Enabled: boolPtr(true),
			},
			expected: true,
		},
		{
			name: "enabled explicitly false",
			cosign: &registry.Cosign{
				Enabled: boolPtr(false),
			},
			expected: false,
		},
		{
			name: "enabled by having opts",
			cosign: &registry.Cosign{
				Opts: []string{"--rekor-url", "https://rekor.sigstore.dev"},
			},
			expected: true,
		},
		{
			name: "enabled by having signature",
			cosign: &registry.Cosign{
				Signature: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: stringPtr("binary.sig"),
				},
			},
			expected: true,
		},
		{
			name: "enabled by having certificate",
			cosign: &registry.Cosign{
				Certificate: &registry.DownloadedFile{
					Type: "http",
					URL:  stringPtr("https://example.com/cert.pem"),
				},
			},
			expected: true,
		},
		{
			name: "enabled by having key",
			cosign: &registry.Cosign{
				Key: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: stringPtr("key.pub"),
				},
			},
			expected: true,
		},
		{
			name: "enabled by having bundle",
			cosign: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: stringPtr("bundle.json"),
				},
			},
			expected: true,
		},
		{
			name: "disabled even with opts when explicitly false",
			cosign: &registry.Cosign{
				Enabled: boolPtr(false),
				Opts:    []string{"--rekor-url", "https://rekor.sigstore.dev"},
				Signature: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: stringPtr("binary.sig"),
				},
			},
			expected: false,
		},
		{
			name: "full configuration enabled",
			cosign: &registry.Cosign{
				Enabled: boolPtr(true),
				Opts:    []string{"--certificate-identity", "test@example.com"},
				Signature: &registry.DownloadedFile{
					Type:      "github_release",
					RepoOwner: "owner",
					RepoName:  "repo",
					Asset:     stringPtr("{{.Asset}}.sig"),
				},
				Certificate: &registry.DownloadedFile{
					Type: "http",
					URL:  stringPtr("https://fulcio.sigstore.dev/api/v1/rootcert"),
				},
			},
			expected: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result := d.cosign.GetEnabled()
			if result != d.expected {
				t.Errorf("expected %v, got %v", d.expected, result)
			}
		})
	}
}

func TestCosign_RenderOpts(t *testing.T) {
	t.Parallel()
	data := []struct {
		name        string
		cosign      *registry.Cosign
		runtime     *runtime.Runtime
		artifact    *template.Artifact
		expected    []string
		expectError bool
	}{
		{
			name: "no opts",
			cosign: &registry.Cosign{
				Enabled: boolPtr(true),
			},
			runtime:  &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			artifact: &template.Artifact{Version: "v1.0.0"},
			expected: []string{},
		},
		{
			name: "simple opts without templates",
			cosign: &registry.Cosign{
				Opts: []string{"--insecure-ignore-tlog", "--insecure-ignore-sct"},
			},
			runtime:  &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			artifact: &template.Artifact{Version: "v1.0.0"},
			expected: []string{"--insecure-ignore-tlog", "--insecure-ignore-sct"},
		},
		{
			name: "opts with version template",
			cosign: &registry.Cosign{
				Opts: []string{"--certificate-identity", "release-{{.Version}}@example.com"},
			},
			runtime: &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			artifact: &template.Artifact{
				Version: "v1.2.3",
			},
			expected: []string{"--certificate-identity", "release-v1.2.3@example.com"},
		},
		{
			name: "opts with multiple templates",
			cosign: &registry.Cosign{
				Opts: []string{
					"--certificate-identity",
					"{{.Version}}@{{.OS}}.example.com",
					"--rekor-url",
					"https://rekor-{{.Arch}}.sigstore.dev",
				},
			},
			runtime: &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			artifact: &template.Artifact{
				Version: "v2.0.0",
				OS:      "Linux",
				Arch:    "x86_64",
			},
			expected: []string{
				"--certificate-identity",
				"v2.0.0@Linux.example.com",
				"--rekor-url",
				"https://rekor-x86_64.sigstore.dev",
			},
		},
		{
			name: "opts with asset template",
			cosign: &registry.Cosign{
				Opts: []string{"--signature", "{{.Asset}}.sig"},
			},
			runtime: &runtime.Runtime{GOOS: "darwin", GOARCH: "arm64"},
			artifact: &template.Artifact{
				Version: "v1.0.0",
				Asset:   "tool_darwin_arm64.tar.gz",
			},
			expected: []string{"--signature", "tool_darwin_arm64.tar.gz.sig"},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			result, err := d.cosign.RenderOpts(d.runtime, d.artifact)

			if d.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(d.expected) {
				t.Errorf("expected %d opts, got %d", len(d.expected), len(result))
				return
			}

			for i, expected := range d.expected {
				if result[i] != expected {
					t.Errorf("expected opt[%d] = %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}
