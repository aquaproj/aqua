package genrgst

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/google/go-cmp/cmp"
)

func strp(s string) *string {
	return &s
}

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

				VersionConstraints: "false",
				VersionOverrides: []*registry.VersionOverride{
					{
						VersionConstraints: "true",
						Asset:              "gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
						Format:             "tar.gz",
						Replacements: registry.Replacements{
							"darwin": "macOS",
						},
						Overrides: []*registry.Override{
							{
								GOOS:   "windows",
								Format: "zip",
							},
						},
						WindowsARMEmulation: ptr.Bool(true),
						Rosetta2:            ptr.Bool(true),
					},
				},
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
				Description: "Fuzzy Finder in rust",
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
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
				Assets:   d.assets,
				Repo:     d.repo,
			}
			cargoClient := &cargo.MockClient{
				CratePayload: d.crate,
			}
			var buf bytes.Buffer
			ctrl := NewController(nil, gh, nil, cargoClient, &buf)
			pkgInfo, _ := ctrl.getPackageInfo(ctx, logger, d.pkgName, &config.Param{}, &Config{})
			if diff := cmp.Diff(d.exp, pkgInfo); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestController_checkChecksumCosign(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name             string
		pkgInfo          *registry.PackageInfo
		checksumFileName string
		assetNames       map[string]struct{}
		want             *registry.Cosign
	}{
		{
			name: "with bundle",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":               {},
				"checksums.txt.cosign.bundle": {},
			},
			want: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: strp("checksums.txt.cosign.bundle"),
				},
				Opts: []string{
					"--certificate-identity-regexp",
					`^https://github\.com/owner/repo/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`,
					"--certificate-oidc-issuer",
					"https://token.actions.githubusercontent.com",
				},
			},
		},
		{
			name: "with certificate and signature",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":             {},
				"checksums.txt-keyless.sig": {},
				"checksums.txt-keyless.pem": {},
			},
			want: &registry.Cosign{
				Opts: []string{
					"--certificate",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.pem",
					"--certificate-identity-regexp",
					`^https://github\.com/owner/repo/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`,
					"--certificate-oidc-issuer",
					"https://token.actions.githubusercontent.com",
					"--signature",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.sig",
				},
			},
		},
		{
			name: "with public key and signature",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":     {},
				"checksums.txt.sig": {},
				"cosign.pub":        {},
			},
			want: &registry.Cosign{
				Opts: []string{
					"--key",
					"https://github.com/owner/repo/releases/download/{{.Version}}/cosign.pub",
					"--signature",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt.sig",
				},
			},
		},
		{
			name: "no cosign files",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt": {},
			},
			want: nil,
		},
		{
			name: "signature only returns nil",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":     {},
				"checksums.txt.sig": {},
			},
			want: nil,
		},
		{
			name: "keyless signature ignores public key",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":             {},
				"checksums.txt-keyless.sig": {},
				"cosign.pub":                {},
			},
			want: nil,
		},
		{
			name: "bundle takes precedence over certificate",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":             {},
				"checksums.txt.bundle":      {},
				"checksums.txt-keyless.sig": {},
				"checksums.txt-keyless.pem": {},
			},
			want: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: strp("checksums.txt.bundle"),
				},
				Opts: []string{
					"--certificate-identity-regexp",
					`^https://github\.com/owner/repo/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`,
					"--certificate-oidc-issuer",
					"https://token.actions.githubusercontent.com",
				},
			},
		},
		{
			name: "bundle takes precedence over public key",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":        {},
				"checksums.txt.bundle": {},
				"checksums.txt.sig":    {},
				"cosign.pub":           {},
			},
			want: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  "github_release",
					Asset: strp("checksums.txt.bundle"),
				},
				Opts: []string{
					"--certificate-identity-regexp",
					`^https://github\.com/owner/repo/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`,
					"--certificate-oidc-issuer",
					"https://token.actions.githubusercontent.com",
				},
			},
		},
		{
			name: "certificate takes precedence over public key",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: "owner",
				RepoName:  "repo",
			},
			checksumFileName: "checksums.txt",
			assetNames: map[string]struct{}{
				"checksums.txt":             {},
				"checksums.txt-keyless.pem": {},
				"checksums.txt-keyless.sig": {},
				"cosign.pub":                {},
			},
			want: &registry.Cosign{
				Opts: []string{
					"--certificate",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.pem",
					"--certificate-identity-regexp",
					`^https://github\.com/owner/repo/\.github/workflows/.+\.ya?ml@refs/tags/\Q{{.Version}}\E$`,
					"--certificate-oidc-issuer",
					"https://token.actions.githubusercontent.com",
					"--signature",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.sig",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := checkChecksumCosign(tt.pkgInfo, tt.checksumFileName, tt.assetNames)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("checkChecksumCosign() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
