//nolint:funlen
package registry_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
	"go.yaml.in/yaml/v3"
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
    cosign:
      enabled: true
    minisign:
      enabled: false
  - version_constraint: ">=v2.0.0"
    asset: package-v2.tar.gz
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

	if vo2.SLSAProvenance == nil {
		t.Fatal("SLSAProvenance should not be nil in second version override")
	}

	if !vo2.SLSAProvenance.GetEnabled() {
		t.Error("SLSAProvenance should be enabled in second version override")
	}
}
