package registry

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const proxyName = "aqua-proxy"

func stringP(s string) *string {
	return &s
}

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
				RepoName:    "colima",
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Asset:       stringP("colima-amd64"),
				Files: []*File{
					{
						Name: proxyName,
					},
				},
			},
			child: &VersionOverride{
				Type: PkgInfoTypeGitHubContent,
				Path: stringP("colima"),
			},
			exp: &PackageInfo{
				Type:        PkgInfoTypeGitHubContent,
				RepoOwner:   "abiosoft",
				RepoName:    "colima",
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Files: []*File{
					{
						Name: proxyName,
					},
				},
				Path: stringP("colima"),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo := d.pkgInfo.overrideVersion(d.child)
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
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
				Type: "github_content",
				Path: stringP("foo"),
			},
			pkgInfo: &PackageInfo{
				Type: "github_content",
				Path: stringP("foo"),
			},
		},
		{
			title: "version constraint",
			exp: &PackageInfo{
				Type:               "github_content",
				Path:               stringP("foo"),
				VersionConstraints: `semver(">= 0.4.0")`,
			},
			pkgInfo: &PackageInfo{
				Type:               "github_content",
				Path:               stringP("foo"),
				VersionConstraints: `semver(">= 0.4.0")`,
			},
			version: "v0.5.0",
		},
		{
			title: "child version constraint",
			exp: &PackageInfo{
				Type:               "github_content",
				Path:               stringP("bar"),
				VersionConstraints: `semver(">= 0.4.0")`,
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: `semver("< 0.4.0")`,
						Path:               stringP("bar"),
					},
				},
			},
			pkgInfo: &PackageInfo{
				Type:               "github_content",
				Path:               stringP("foo"),
				VersionConstraints: `semver(">= 0.4.0")`,
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: `semver("< 0.4.0")`,
						Path:               stringP("bar"),
					},
				},
			},
			version: "v0.3.0",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, err := d.pkgInfo.SetVersion(d.version)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(pkgInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
