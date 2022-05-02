package config

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/template"
	constraint "github.com/aquaproj/aqua/pkg/version-constraint"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
				Asset:       template.NewTemplate("colima-amd64"),
				Files: []*File{
					{
						Name: proxyName,
					},
				},
			},
			child: &VersionOverride{
				Type: PkgInfoTypeGitHubContent,
				Path: template.NewTemplate("colima"),
			},
			exp: &PackageInfo{
				Type:        PkgInfoTypeGitHubContent,
				RepoOwner:   "abiosoft",
				RepoName:    "colima",
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Asset:       template.NewTemplate("colima-amd64"),
				Files: []*File{
					{
						Name: proxyName,
					},
				},
				Path: template.NewTemplate("colima"),
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo := d.pkgInfo.overrideVersion(d.child)
			if diff := cmp.Diff(d.exp, pkgInfo, cmp.AllowUnexported(template.Template{})); diff != "" {
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
				Path: template.NewTemplate("foo"),
			},
			pkgInfo: &PackageInfo{
				Type: "github_content",
				Path: template.NewTemplate("foo"),
			},
		},
		{
			title: "version constraint",
			exp: &PackageInfo{
				Type:               "github_content",
				Path:               template.NewTemplate("foo"),
				VersionConstraints: constraint.NewVersionConstraints(`semver(">= 0.4.0")`),
			},
			pkgInfo: &PackageInfo{
				Type:               "github_content",
				Path:               template.NewTemplate("foo"),
				VersionConstraints: constraint.NewVersionConstraints(`semver(">= 0.4.0")`),
			},
			version: "v0.5.0",
		},
		{
			title: "child version constraint",
			exp: &PackageInfo{
				Type: "github_content",
				Path: template.NewTemplate("bar"),
			},
			pkgInfo: &PackageInfo{
				Type:               "github_content",
				Path:               template.NewTemplate("foo"),
				VersionConstraints: constraint.NewVersionConstraints(`semver(">= 0.4.0")`),
				VersionOverrides: []*VersionOverride{
					{
						VersionConstraints: constraint.NewVersionConstraints(`semver("< 0.4.0")`),
						Path:               template.NewTemplate("bar"),
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
			pkgInfo, err := d.pkgInfo.setVersion(d.version)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(pkgInfo, d.exp, cmpopts.IgnoreUnexported(constraint.VersionConstraints{}), cmp.AllowUnexported(template.Template{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
