package genrgst

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestController_getPackageInfo(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		pkgName  string
		exp      *registry.PackageInfo
		releases []*github.RepositoryRelease
		repo     *github.Repository
		assets   []*github.ReleaseAsset
		crate    *cargo.CratePayload
	}{
		{
			name:    "package name doesn't have slash",
			pkgName: "foo",
			exp: &registry.PackageInfo{
				Name: "foo",
				Type: "github_release",
			},
		},
		{
			name:    "repo not found",
			pkgName: "foo/foo",
			exp: &registry.PackageInfo{
				RepoOwner: "foo",
				RepoName:  "foo",
				Type:      "github_release",
			},
		},
		{
			name:    "no release",
			pkgName: "foo/foo",
			exp: &registry.PackageInfo{
				RepoOwner:   "foo",
				RepoName:    "foo",
				Type:        "github_release",
				Description: "hello",
			},
			repo: &github.Repository{
				Description: ptr.String("hello."),
			},
		},
		{
			name:    "normal",
			pkgName: "cli/cli",
			exp: &registry.PackageInfo{
				RepoOwner:   "cli",
				RepoName:    "cli",
				Type:        "github_release",
				Description: "GitHub’s official command line tool",
				Asset:       "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
				Format:      "tar.gz",
				Replacements: registry.Replacements{
					"darwin": "macOS",
				},
				Overrides: []*registry.Override{
					{
						GOOS:   "windows",
						Format: "zip",
					},
				},
				SupportedEnvs: registry.SupportedEnvs{
					"darwin",
					"linux",
					"amd64",
				},
				Rosetta2: true,
			},
			repo: &github.Repository{
				Description: ptr.String("GitHub’s official command line tool"),
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: ptr.String("v2.13.0"),
				},
			},
			assets: []*github.ReleaseAsset{
				{
					Name: ptr.String("gh_2.13.0_linux_amd64.tar.gz"),
				},
				{
					Name: ptr.String("gh_2.13.0_linux_arm64.tar.gz"),
				},
				{
					Name: ptr.String("gh_2.13.0_macOS_amd64.tar.gz"),
				},
				{
					Name: ptr.String("gh_2.13.0_windows_amd64.zip"),
				},
			},
		},
		{
			name:    "cargo",
			pkgName: "crates.io/skim",
			exp: &registry.PackageInfo{
				Name:        "crates.io/skim",
				RepoOwner:   "lotabout",
				RepoName:    "skim",
				Type:        "cargo",
				Crate:       "skim",
				Description: "Fuzzy Finder in rust!",
			},
			crate: &cargo.CratePayload{
				Crate: &cargo.Crate{
					Homepage:    "https://github.com/lotabout/skim",
					Description: "Fuzzy Finder in rust!",
					Repository:  "https://github.com/lotabout/skim",
				},
			},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
				Assets:   d.assets,
				Repo:     d.repo,
			}
			cargoClient := &cargo.MockClient{
				CratePayload: d.crate,
			}
			ctrl := NewController(nil, gh, nil, cargoClient)
			pkgInfo, _ := ctrl.getPackageInfo(ctx, logE, d.pkgName, &config.Param{
				Type: "github_release",
				Deep: true,
			})
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
