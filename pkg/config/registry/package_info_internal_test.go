package registry

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
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
				RepoName:    "colima",
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
				Path: "colima",
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
				Path: "colima",
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
				Path: "foo",
			},
			pkgInfo: &PackageInfo{
				Type: "github_content",
				Path: "foo",
			},
		},
		{
			title: "version constraint",
			exp: &PackageInfo{
				Type:               "github_content",
				Path:               "foo",
				VersionConstraints: `semver(">= 0.4.0")`,
			},
			pkgInfo: &PackageInfo{
				Type:               "github_content",
				Path:               "foo",
				VersionConstraints: `semver(">= 0.4.0")`,
			},
			version: "v0.5.0",
		},
		{
			title: "child version constraint",
			exp: &PackageInfo{
				Type:               "github_content",
				Path:               "bar",
				VersionConstraints: `semver(">= 0.4.0")`,
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: `semver("< 0.4.0")`,
						Path:               "bar",
					},
				},
			},
			pkgInfo: &PackageInfo{
				Type:               "github_content",
				Path:               "foo",
				VersionConstraints: `semver(">= 0.4.0")`,
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: `semver("< 0.4.0")`,
						Path:               "bar",
					},
				},
			},
			version: "v0.3.0",
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo, err := d.pkgInfo.SetVersion(logE, d.version)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(pkgInfo, d.exp); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
