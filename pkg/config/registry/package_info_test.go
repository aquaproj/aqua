//nolint:funlen
package registry_test

import (
	"encoding/json"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
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

func TestPackageInfo_GetFiles(t *testing.T) {
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

func TestPackageInfo_MaybeHasCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		has     []string
		lacks   []string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			has: []string{
				"go", "gofmt", "go.build", "go.override", "go.v", "go.vbuild", "go.voverride",
			},
			lacks: []string{"golang"},
			pkgInfo: &registry.PackageInfo{
				RepoName: "golang",
				Files: []*registry.File{
					{
						Name: "go",
					},
					{
						Name: "gofmt",
					},
				},
				Build: &registry.Build{
					Files: []*registry.File{
						{
							Name: "go.build",
						},
					},
				},
				Overrides: []*registry.Override{
					{
						Files: []*registry.File{
							{
								Name: "go.override",
							},
						},
					},
				},
				VersionOverrides: []*registry.VersionOverride{
					{
						Files: []*registry.File{
							{
								Name: "go.v",
							},
						},
						Build: &registry.Build{
							Files: []*registry.File{
								{
									Name: "go.vbuild",
								},
							},
						},
						Overrides: []*registry.Override{
							{
								Files: []*registry.File{
									{
										Name: "go.voverride",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			title: "empty",
			has:   []string{"ci-info"},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "potentially empty",
			has:   []string{"ci-info", "ci-info.prebuilt"},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Files: []*registry.File{
					{
						Name: "ci-info.prebuilt",
					},
				},
				Build: &registry.Build{},
			},
		},
		{
			title: "has name",
			has:   []string{"cmctl"},
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
			for _, exe := range d.has {
				if !d.pkgInfo.MaybeHasCommand(exe) {
					t.Fatalf("expected to have command %s", exe)
				}
			}
			for _, exe := range d.lacks {
				if d.pkgInfo.MaybeHasCommand(exe) {
					t.Fatalf("expected to not have command %s", exe)
				}
			}
		})
	}
}

func TestPackageInfo_Validate(t *testing.T) {
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

func TestPackageInfo_JSONEncode_VersionOverrides_ImmutableRelease(t *testing.T) { //nolint:cyclop
	t.Parallel()

	// Helper function for bool pointer
	boolPtr := func(b bool) *bool { return &b }

	pkgInfo := &registry.PackageInfo{
		Name:      "test-package",
		Type:      "github_release",
		RepoOwner: "owner",
		RepoName:  "repo",
		Asset:     "asset.tar.gz",
		VersionOverrides: []*registry.VersionOverride{
			{
				VersionConstraints:     ">=v1.0.0",
				GitHubImmutableRelease: boolPtr(true),
			},
			{
				VersionConstraints:     ">=v2.0.0",
				GitHubImmutableRelease: boolPtr(false),
			},
			{
				VersionConstraints: ">=v3.0.0",
				// ImmutableRelease with nil (should be omitted due to omitempty)
				GitHubImmutableRelease: nil,
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

	// Test first version override (github_immutable_release: true)
	vo1, ok := versionOverrides[0].(map[string]any)
	if !ok {
		t.Fatal("first version override is not a map")
	}

	immutableRelease1, ok := vo1["github_immutable_release"].(bool)
	if !ok || !immutableRelease1 {
		t.Errorf("expected github_immutable_release: true in first version override, got %v", vo1["github_immutable_release"])
	}

	// Test second version override (github_immutable_release: false)
	vo2, ok := versionOverrides[1].(map[string]any)
	if !ok {
		t.Fatal("second version override is not a map")
	}

	immutableRelease2, ok := vo2["github_immutable_release"].(bool)
	if !ok || immutableRelease2 {
		t.Errorf("expected github_immutable_release: false in second version override, got %v", vo2["github_immutable_release"])
	}

	// Test third version override (github_immutable_release: nil - should be omitted due to omitempty)
	vo3, ok := versionOverrides[2].(map[string]any)
	if !ok {
		t.Fatal("third version override is not a map")
	}

	// With omitempty, nil github_immutable_release should be omitted from JSON
	if _, exists := vo3["github_immutable_release"]; exists {
		t.Errorf("expected github_immutable_release field to be omitted in third version override, but found: %v", vo3["github_immutable_release"])
	}
}

func TestPackageInfo_JSONEncode_RoundTrip(t *testing.T) {
	t.Parallel()

	// Helper function for bool pointer
	boolPtr := func(b bool) *bool { return &b }

	original := &registry.PackageInfo{
		Name:      "kubectl",
		Type:      "github_release",
		RepoOwner: "kubernetes",
		RepoName:  "kubernetes",
		Asset:     "kubectl",
		VersionOverrides: []*registry.VersionOverride{
			{
				VersionConstraints:     ">=v1.25.0",
				Asset:                  "kubectl-new",
				GitHubImmutableRelease: boolPtr(true),
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
	if vo.GitHubImmutableRelease == nil {
		t.Fatal("ImmutableRelease should not be nil after JSON round trip")
	}

	if !*vo.GitHubImmutableRelease {
		t.Error("ImmutableRelease should be true after JSON round trip")
	}

	// Compare original and decoded using deep comparison
	if diff := cmp.Diff(*original.VersionOverrides[0].GitHubImmutableRelease,
		*decoded.VersionOverrides[0].GitHubImmutableRelease); diff != "" {
		t.Errorf("ImmutableRelease value differs after JSON round trip (-want +got):\n%s", diff)
	}
}

func TestPackageInfo_YAMLDecode_VersionOverrides_ImmutableRelease(t *testing.T) { //nolint:cyclop
	t.Parallel()

	// YAML with VersionOverrides containing ImmutableRelease
	yamlData := `
name: test-package
type: github_release
repo_owner: owner
repo_name: repo
asset: asset.tar.gz
version_overrides:
  - version_constraint: ">=v1.0.0"
    github_immutable_release: true
  - version_constraint: ">=v2.0.0"
    github_immutable_release: false
  - version_constraint: ">=v3.0.0"
    # No github_immutable_release field (should be nil)
`

	var pkgInfo registry.PackageInfo
	if err := yaml.Unmarshal([]byte(yamlData), &pkgInfo); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// Verify basic package info
	if pkgInfo.Name != "test-package" {
		t.Errorf("expected name 'test-package', got %q", pkgInfo.Name)
	}

	if pkgInfo.Type != "github_release" {
		t.Errorf("expected type 'github_release', got %q", pkgInfo.Type)
	}

	// Verify version overrides
	if len(pkgInfo.VersionOverrides) != 3 {
		t.Fatalf("expected 3 version overrides, got %d", len(pkgInfo.VersionOverrides))
	}

	// Test first version override (github_immutable_release: true)
	vo1 := pkgInfo.VersionOverrides[0]
	if vo1.VersionConstraints != ">=v1.0.0" {
		t.Errorf("expected version constraint '>=v1.0.0', got %q", vo1.VersionConstraints)
	}

	if vo1.GitHubImmutableRelease == nil {
		t.Fatal("ImmutableRelease should not be nil in first version override")
	}

	if !*vo1.GitHubImmutableRelease {
		t.Error("ImmutableRelease should be true in first version override")
	}

	// Test second version override (github_immutable_release: false)
	vo2 := pkgInfo.VersionOverrides[1]
	if vo2.VersionConstraints != ">=v2.0.0" {
		t.Errorf("expected version constraint '>=v2.0.0', got %q", vo2.VersionConstraints)
	}

	if vo2.GitHubImmutableRelease == nil {
		t.Fatal("ImmutableRelease should not be nil in second version override")
	}

	if *vo2.GitHubImmutableRelease {
		t.Error("ImmutableRelease should be false in second version override")
	}

	// Test third version override (github_immutable_release: nil)
	vo3 := pkgInfo.VersionOverrides[2]
	if vo3.VersionConstraints != ">=v3.0.0" {
		t.Errorf("expected version constraint '>=v3.0.0', got %q", vo3.VersionConstraints)
	}

	if vo3.GitHubImmutableRelease != nil {
		t.Errorf("ImmutableRelease should be nil in third version override, got %v", *vo3.GitHubImmutableRelease)
	}
}

func TestPackageInfo_YAMLDecode_RoundTrip(t *testing.T) {
	t.Parallel()

	// Helper function for bool pointer
	boolPtr := func(b bool) *bool { return &b }

	original := &registry.PackageInfo{
		Name:      "kubectl",
		Type:      "github_release",
		RepoOwner: "kubernetes",
		RepoName:  "kubernetes",
		Asset:     "kubectl",
		VersionOverrides: []*registry.VersionOverride{
			{
				VersionConstraints:     ">=v1.25.0",
				Asset:                  "kubectl-new",
				GitHubImmutableRelease: boolPtr(true),
			},
		},
	}

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal PackageInfo to YAML: %v", err)
	}

	// Unmarshal back to PackageInfo
	var decoded registry.PackageInfo
	if err := yaml.Unmarshal(yamlBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	// Verify the decoded GitHubReleaseAttestation is correctly preserved
	if len(decoded.VersionOverrides) != 1 {
		t.Fatalf("expected 1 version override, got %d", len(decoded.VersionOverrides))
	}

	vo := decoded.VersionOverrides[0]
	if vo.GitHubImmutableRelease == nil {
		t.Fatal("ImmutableRelease should not be nil after YAML round trip")
	}

	if !*vo.GitHubImmutableRelease {
		t.Error("ImmutableRelease should be true after YAML round trip")
	}

	// Compare original and decoded using deep comparison
	if diff := cmp.Diff(*original.VersionOverrides[0].GitHubImmutableRelease,
		*decoded.VersionOverrides[0].GitHubImmutableRelease); diff != "" {
		t.Errorf("ImmutableRelease value differs after YAML round trip (-want +got):\n%s", diff)
	}

	// Verify other fields are preserved
	if decoded.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, decoded.Name)
	}

	if decoded.VersionOverrides[0].Asset != original.VersionOverrides[0].Asset {
		t.Errorf("expected asset %q, got %q", original.VersionOverrides[0].Asset, decoded.VersionOverrides[0].Asset)
	}
}

func TestPackageInfo_YAMLDecode_NestedStructure(t *testing.T) { //nolint:cyclop
	t.Parallel()

	// Test more complex YAML structure with multiple verification configurations
	yamlData := `
name: complex-package
type: github_release
repo_owner: owner
repo_name: repo
asset: package.tar.gz
version_overrides:
  - version_constraint: ">=v1.0.0"
    asset: package-v1.tar.gz
    github_immutable_release: true
    cosign:
      enabled: true
    minisign:
      enabled: false
  - version_constraint: ">=v2.0.0"
    asset: package-v2.tar.gz
    github_immutable_release: false
    slsa_provenance:
      enabled: true
      type: github_release
`

	var pkgInfo registry.PackageInfo
	if err := yaml.Unmarshal([]byte(yamlData), &pkgInfo); err != nil {
		t.Fatalf("failed to unmarshal complex YAML: %v", err)
	}

	// Verify version overrides
	if len(pkgInfo.VersionOverrides) != 2 {
		t.Fatalf("expected 2 version overrides, got %d", len(pkgInfo.VersionOverrides))
	}

	// Test first version override
	vo1 := pkgInfo.VersionOverrides[0]
	if vo1.Asset != "package-v1.tar.gz" {
		t.Errorf("expected asset 'package-v1.tar.gz', got %q", vo1.Asset)
	}

	if vo1.GitHubImmutableRelease == nil {
		t.Fatal("ImmutableRelease should not be nil in first version override")
	}

	if !*vo1.GitHubImmutableRelease {
		t.Error("ImmutableRelease should be true in first version override")
	}

	if vo1.Cosign == nil {
		t.Fatal("Cosign should not be nil in first version override")
	}

	if !vo1.Cosign.GetEnabled() {
		t.Error("Cosign should be enabled in first version override")
	}

	if vo1.Minisign == nil {
		t.Fatal("Minisign should not be nil in first version override")
	}

	if vo1.Minisign.GetEnabled() {
		t.Error("Minisign should be disabled in first version override")
	}

	// Test second version override
	vo2 := pkgInfo.VersionOverrides[1]
	if vo2.Asset != "package-v2.tar.gz" {
		t.Errorf("expected asset 'package-v2.tar.gz', got %q", vo2.Asset)
	}

	if vo2.GitHubImmutableRelease == nil {
		t.Fatal("ImmutableRelease should not be nil in second version override")
	}

	if *vo2.GitHubImmutableRelease {
		t.Error("ImmutableRelease should be false in second version override")
	}

	if vo2.SLSAProvenance == nil {
		t.Fatal("SLSAProvenance should not be nil in second version override")
	}

	if !vo2.SLSAProvenance.GetEnabled() {
		t.Error("SLSAProvenance should be enabled in second version override")
	}
}
