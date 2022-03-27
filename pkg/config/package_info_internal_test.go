package config

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/template"
	"github.com/google/go-cmp/cmp"
)

const proxyName = "aqua-proxy"

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
			child: &PackageInfo{
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
			d.pkgInfo.override(d.child)
			if diff := cmp.Diff(d.exp, d.pkgInfo, cmp.AllowUnexported(template.Template{})); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
