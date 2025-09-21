package registry_test

import (
	"encoding/json"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestPackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Name: "foo",
			},
		},
		{
			title: "default",
			exp:   "suzuki-shunsuke/ci-info",
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if name := d.pkgInfo.GetName(); name != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, name)
			}
		})
	}
}

func TestPackageInfo_GetLink(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Link: "http://example.com",
			},
		},
		{
			title: "default",
			exp:   "https://github.com/suzuki-shunsuke/ci-info",
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if link := d.pkgInfo.GetLink(); link != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, link)
			}
		})
	}
}

func TestPackageInfo_GetFormat(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &registry.PackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &registry.PackageInfo{},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if format := d.pkgInfo.GetFormat(); format != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, format)
			}
		})
	}
}

func TestPackageInfo_GetFiles(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		exp     []*registry.File
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp: []*registry.File{
				{
					Name: "go",
				},
				{
					Name: "gofmt",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Files: []*registry.File{
					{
						Name: "go",
					},
					{
						Name: "gofmt",
					},
				},
			},
		},
		{
			title: "empty",
			exp: []*registry.File{
				{
					Name: "ci-info",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "has name",
			exp: []*registry.File{
				{
					Name: "cmctl",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "cert-manager",
				RepoName:  "cert-manager",
				Name:      "cert-manager/cert-manager/cmctl",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			files := d.pkgInfo.GetFiles()
			if diff := cmp.Diff(d.exp, files); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPackageInfo_Validate(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title:   "package name is required",
			pkgInfo: &registry.PackageInfo{},
			isErr:   true,
		},
		{
			title: "repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubArchive,
			},
			isErr: true,
		},
		{
			title: "github_archive",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubArchive,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "github_content repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubContent,
			},
			isErr: true,
		},
		{
			title: "github_content path is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubContent,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
			isErr: true,
		},
		{
			title: "github_content",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubContent,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Path:      "bin/ci-info",
			},
		},
		{
			title: "github_release repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubRelease,
			},
			isErr: true,
		},
		{
			title: "github_release asset is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
			isErr: true,
		},
		{
			title: "github_release",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     "ci-info.tar.gz",
			},
		},
		{
			title: "http url is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeHTTP,
			},
			isErr: true,
		},
		{
			title: "http",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeHTTP,
				Name: "suzuki-shunsuke/ci-info",
				URL:  "http://example.com",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
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

func TestPackageInfo_JSONEncode_VersionOverrides_GitHubReleaseAttestation(t *testing.T) { //nolint:cyclop
	t.Parallel()

	pkgInfo := &registry.PackageInfo{
		Name:      "test-package",
		Type:      "github_release",
		RepoOwner: "owner",
		RepoName:  "repo",
		Asset:     "asset.tar.gz",
		VersionOverrides: []*registry.VersionOverride{
			{
				VersionConstraints: ">=v1.0.0",
				GitHubReleaseAttestation: &registry.GitHubReleaseAttestation{
					Enabled: boolPtr(true),
				},
			},
			{
				VersionConstraints: ">=v2.0.0",
				GitHubReleaseAttestation: &registry.GitHubReleaseAttestation{
					Enabled: boolPtr(false),
				},
			},
			{
				VersionConstraints: ">=v3.0.0",
				// GitHubReleaseAttestation with nil Enabled (should default to true)
				GitHubReleaseAttestation: &registry.GitHubReleaseAttestation{},
			},
		},
	}

	// Test JSON encoding
	jsonBytes, err := json.Marshal(pkgInfo)
	if err != nil {
		t.Fatalf("failed to marshal PackageInfo to JSON: %v", err)
	}

	// Verify JSON contains the expected GitHubReleaseAttestation fields
	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Check that version_overrides is present
	versionOverrides, ok := result["version_overrides"].([]any)
	if !ok {
		t.Fatal("version_overrides not found or not an array")
	}

	if len(versionOverrides) != 3 {
		t.Fatalf("expected 3 version overrides, got %d", len(versionOverrides))
	}

	// Test first version override (enabled: true)
	vo1, ok := versionOverrides[0].(map[string]any)
	if !ok {
		t.Fatal("first version override is not a map")
	}

	gra1, ok := vo1["github_release_attestation"].(map[string]any)
	if !ok {
		t.Fatal("github_release_attestation not found in first version override")
	}

	enabled1, ok := gra1["enabled"].(bool)
	if !ok || !enabled1 {
		t.Errorf("expected enabled: true in first version override, got %v", gra1["enabled"])
	}

	// Test second version override (enabled: false)
	vo2, ok := versionOverrides[1].(map[string]any)
	if !ok {
		t.Fatal("second version override is not a map")
	}

	gra2, ok := vo2["github_release_attestation"].(map[string]any)
	if !ok {
		t.Fatal("github_release_attestation not found in second version override")
	}

	enabled2, ok := gra2["enabled"].(bool)
	if !ok || enabled2 {
		t.Errorf("expected enabled: false in second version override, got %v", gra2["enabled"])
	}

	// Test third version override (enabled: nil - should be omitted due to omitempty)
	vo3, ok := versionOverrides[2].(map[string]any)
	if !ok {
		t.Fatal("third version override is not a map")
	}

	gra3, ok := vo3["github_release_attestation"].(map[string]any)
	if !ok {
		t.Fatal("github_release_attestation not found in third version override")
	}

	// With omitempty, nil enabled should be omitted from JSON
	if _, exists := gra3["enabled"]; exists {
		t.Errorf("expected enabled field to be omitted in third version override, but found: %v", gra3["enabled"])
	}
}

func TestPackageInfo_JSONEncode_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &registry.PackageInfo{
		Name:      "kubectl",
		Type:      "github_release",
		RepoOwner: "kubernetes",
		RepoName:  "kubernetes",
		Asset:     "kubectl",
		VersionOverrides: []*registry.VersionOverride{
			{
				VersionConstraints: ">=v1.25.0",
				Asset:              "kubectl-new",
				GitHubReleaseAttestation: &registry.GitHubReleaseAttestation{
					Enabled: boolPtr(true),
				},
			},
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal PackageInfo: %v", err)
	}

	// Unmarshal back to PackageInfo
	var decoded registry.PackageInfo
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal PackageInfo: %v", err)
	}

	// Verify the decoded GitHubReleaseAttestation is correctly preserved
	if len(decoded.VersionOverrides) != 1 {
		t.Fatalf("expected 1 version override, got %d", len(decoded.VersionOverrides))
	}

	vo := decoded.VersionOverrides[0]
	if vo.GitHubReleaseAttestation == nil {
		t.Fatal("GitHubReleaseAttestation should not be nil after JSON round trip")
	}

	if !vo.GitHubReleaseAttestation.GetEnabled() {
		t.Error("GitHubReleaseAttestation should be enabled after JSON round trip")
	}

	// Compare original and decoded using deep comparison
	if diff := cmp.Diff(original.VersionOverrides[0].GitHubReleaseAttestation.GetEnabled(),
		decoded.VersionOverrides[0].GitHubReleaseAttestation.GetEnabled()); diff != "" {
		t.Errorf("GitHubReleaseAttestation enabled status differs after JSON round trip (-want +got):\n%s", diff)
	}
}
