package controller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPackageInfo_override(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *PackageInfo
		child   *PackageInfo
		exp     *PackageInfo
	}{
		{
			title: "normal",
			pkgInfo: &PackageInfo{
				Type:        pkgInfoTypeGitHubRelease,
				RepoOwner:   "abiosoft",
				RepoName:    "colima",
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Asset: &Template{
					raw: "colima-amd64",
				},
				Files: []*File{
					{
						Name: proxyName,
					},
				},
			},
			child: &PackageInfo{
				Type: pkgInfoTypeGitHubContent,
				Path: &Template{
					raw: "colima",
				},
			},
			exp: &PackageInfo{
				Type:        pkgInfoTypeGitHubContent,
				RepoOwner:   "abiosoft",
				RepoName:    "colima",
				Description: "Docker (and Kubernetes) on MacOS with minimal setup",
				Asset: &Template{
					raw: "colima-amd64",
				},
				Files: []*File{
					{
						Name: proxyName,
					},
				},
				Path: &Template{
					raw: "colima",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			d.pkgInfo.override(d.child)
			if diff := cmp.Diff(d.exp, d.pkgInfo, cmp.AllowUnexported(Template{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
