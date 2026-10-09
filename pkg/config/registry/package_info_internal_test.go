package registry

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const proxyName = "aqua-proxy"

func TestPackageInfo_overrideVersion(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *PackageInfo
		child   *VersionOverride
		exp     *PackageInfo
	}{
		{
			title: "normal",
			pkgInfo: &PackageInfo{
				Type:        PkgInfoTypeGitHubRelease,
				RepoOwner:   "abiosoft",
				RepoName:    repoNameColima,
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Asset:       "colima-amd64",
				Files: []*File{
					{
						Name: proxyName,
					},
				},
			},
			child: &VersionOverride{
				Type: PkgInfoTypeGitHubContent,
				Path: repoNameColima,
			},
			exp: &PackageInfo{
				Type:        PkgInfoTypeGitHubContent,
				RepoOwner:   "abiosoft",
				RepoName:    repoNameColima,
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Files: []*File{
					{
						Name: proxyName,
					},
				},
				Path: repoNameColima,
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo := d.pkgInfo.overrideVersion(d.child)
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

// A project that moved to a Forgejo instance keeps github_release at the top level and
// says forgejo_release for the newer versions. What belonged to GitHub has to go with it:
// the override can't unset a field, because a field it doesn't mention is inherited.
func TestPackageInfo_overrideVersion_toForgeRelease(t *testing.T) {
	t.Parallel()
	pkgInfo := &PackageInfo{
		Type:                       PkgInfoTypeGitHubRelease,
		RepoOwner:                  "an-owner",
		RepoName:                   "a-repo",
		Asset:                      "a-repo.tar.gz",
		Private:                    true,
		SLSAProvenance:             &SLSAProvenance{},
		GitHubArtifactAttestations: &GitHubArtifactAttestations{},
	}
	pkg := pkgInfo.overrideVersion(&VersionOverride{
		Type: PkgInfoTypeForgejoRelease,
		Host: "codeberg.org",
	})
	if pkg.SLSAProvenance != nil || pkg.GitHubArtifactAttestations != nil || pkg.Private {
		t.Fatalf("wanted what only GitHub can answer to be gone, got %+v", pkg)
	}
	if pkg.Host != "codeberg.org" {
		t.Fatalf("wanted the instance the override named, got %s", pkg.Host)
	}
	if err := pkg.Validate(); err != nil {
		t.Fatal(err)
	}
}

// The other way round, the instance goes: a github_release package is on github.com.
func TestPackageInfo_overrideVersion_fromForgeRelease(t *testing.T) {
	t.Parallel()
	pkgInfo := &PackageInfo{
		Type:      PkgInfoTypeForgejoRelease,
		Host:      "codeberg.org",
		RepoOwner: "an-owner",
		RepoName:  "a-repo",
		Asset:     "a-repo.tar.gz",
	}
	pkg := pkgInfo.overrideVersion(&VersionOverride{
		Type: PkgInfoTypeGitHubRelease,
	})
	if pkg.Host != "" {
		t.Fatalf("wanted no instance, got %s", pkg.Host)
	}
}

func TestPackageInfo_setVersion(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		version string
		pkgInfo *PackageInfo
		exp     *PackageInfo
	}{
		{
			title: "no version constraint",
			exp: &PackageInfo{
				Type: PkgInfoTypeGitHubContent,
				Path: pkgFoo,
			},
			pkgInfo: &PackageInfo{
				Type: PkgInfoTypeGitHubContent,
				Path: pkgFoo,
			},
		},
		{
			title: "version constraint",
			exp: &PackageInfo{
				Type:               PkgInfoTypeGitHubContent,
				Path:               pkgFoo,
				VersionConstraints: semverGTE040,
			},
			pkgInfo: &PackageInfo{
				Type:               PkgInfoTypeGitHubContent,
				Path:               pkgFoo,
				VersionConstraints: semverGTE040,
			},
			version: "v0.5.0",
		},
		{
			title: "child version constraint",
			exp: &PackageInfo{
				Type:               PkgInfoTypeGitHubContent,
				Path:               pkgBar,
				VersionConstraints: semverGTE040,
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: `semver("< 0.4.0")`,
						Path:               pkgBar,
					},
				},
			},
			pkgInfo: &PackageInfo{
				Type:               PkgInfoTypeGitHubContent,
				Path:               pkgFoo,
				VersionConstraints: semverGTE040,
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: `semver("< 0.4.0")`,
						Path:               pkgBar,
					},
				},
			},
			version: "v0.3.0",
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, err := d.pkgInfo.SetVersion(logger, d.version)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(pkgInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
