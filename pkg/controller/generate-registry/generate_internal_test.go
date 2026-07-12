package genrgst

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/cargo"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/google/go-cmp/cmp"
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
			pkgName: pkgFoo,
			exp: &registry.PackageInfo{
				Name: pkgFoo,
				Type: pkgTypeGitHubRelease,
			},
		},
		{
			name:    "repo not found",
			pkgName: "foo/foo",
			exp: &registry.PackageInfo{
				RepoOwner: pkgFoo,
				RepoName:  pkgFoo,
				Type:      pkgTypeGitHubRelease,
			},
		},
		{
			name:    "no release",
			pkgName: "foo/foo",
			exp: &registry.PackageInfo{
				RepoOwner:   pkgFoo,
				RepoName:    pkgFoo,
				Type:        pkgTypeGitHubRelease,
				Description: "hello",
			},
			repo: &github.Repository{
				Description: new("hello."),
			},
		},
		{
			name:    caseNormal,
			pkgName: "cli/cli",
			exp: &registry.PackageInfo{
				RepoOwner:   "cli",
				RepoName:    "cli",
				Type:        pkgTypeGitHubRelease,
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
						WindowsARMEmulation: new(true),
						Rosetta2:            new(true),
					},
				},
			},
			repo: &github.Repository{
				Description: new("GitHub’s official command line tool"),
			},
			releases: []*github.RepositoryRelease{
				{
					TagName: "v2.13.0",
				},
			},
			assets: []*github.ReleaseAsset{
				{
					Name: new("gh_2.13.0_linux_amd64.tar.gz"),
				},
				{
					Name: new("gh_2.13.0_linux_arm64.tar.gz"),
				},
				{
					Name: new("gh_2.13.0_macOS_amd64.tar.gz"),
				},
				{
					Name: new("gh_2.13.0_windows_amd64.zip"),
				},
			},
		},
		{
			name:    pkgTypeCargo,
			pkgName: "crates.io/skim",
			exp: &registry.PackageInfo{
				Name:        "crates.io/skim",
				RepoOwner:   "lotabout",
				RepoName:    "skim",
				Type:        pkgTypeCargo,
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
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:              {},
				"checksums.txt.cosign.bundle": {},
			},
			want: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  pkgTypeGitHubRelease,
					Asset: new("checksums.txt.cosign.bundle"),
				},
				Opts: []string{
					flagCertIdentityRegexp,
					regexpCertIdentity,
					flagCertOIDCIssuer,
					urlOIDCIssuer,
				},
			},
		},
		{
			name: "with certificate and signature",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:      {},
				fileChecksumsKeyless:  {},
				fileChecksumsKeylessP: {},
			},
			want: &registry.Cosign{
				Opts: []string{
					"--certificate",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.pem",
					flagCertIdentityRegexp,
					regexpCertIdentity,
					flagCertOIDCIssuer,
					urlOIDCIssuer,
					flagSignature,
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.sig",
				},
			},
		},
		{
			name: "with public key and signature",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:    {},
				fileChecksumsTxtSig: {},
				fileCosignPub:       {},
			},
			want: &registry.Cosign{
				Opts: []string{
					"--key",
					"https://github.com/owner/repo/releases/download/{{.Version}}/cosign.pub",
					flagSignature,
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt.sig",
				},
			},
		},
		{
			name: "no cosign files",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt: {},
			},
			want: nil,
		},
		{
			name: "signature only returns nil",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:    {},
				fileChecksumsTxtSig: {},
			},
			want: nil,
		},
		{
			name: "keyless signature ignores public key",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:     {},
				fileChecksumsKeyless: {},
				fileCosignPub:        {},
			},
			want: nil,
		},
		{
			name: "bundle takes precedence over certificate",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:       {},
				"checksums.txt.bundle": {},
				fileChecksumsKeyless:   {},
				fileChecksumsKeylessP:  {},
			},
			want: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  pkgTypeGitHubRelease,
					Asset: new("checksums.txt.bundle"),
				},
				Opts: []string{
					flagCertIdentityRegexp,
					regexpCertIdentity,
					flagCertOIDCIssuer,
					urlOIDCIssuer,
				},
			},
		},
		{
			name: "bundle takes precedence over public key",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:       {},
				"checksums.txt.bundle": {},
				fileChecksumsTxtSig:    {},
				fileCosignPub:          {},
			},
			want: &registry.Cosign{
				Bundle: &registry.DownloadedFile{
					Type:  pkgTypeGitHubRelease,
					Asset: new("checksums.txt.bundle"),
				},
				Opts: []string{
					flagCertIdentityRegexp,
					regexpCertIdentity,
					flagCertOIDCIssuer,
					urlOIDCIssuer,
				},
			},
		},
		{
			name: "certificate takes precedence over public key",
			pkgInfo: &registry.PackageInfo{
				RepoOwner: repoOwner,
				RepoName:  repoName,
			},
			checksumFileName: fileChecksumsTxt,
			assetNames: map[string]struct{}{
				fileChecksumsTxt:      {},
				fileChecksumsKeylessP: {},
				fileChecksumsKeyless:  {},
				fileCosignPub:         {},
			},
			want: &registry.Cosign{
				Opts: []string{
					"--certificate",
					"https://github.com/owner/repo/releases/download/{{.Version}}/checksums.txt-keyless.pem",
					flagCertIdentityRegexp,
					regexpCertIdentity,
					flagCertOIDCIssuer,
					urlOIDCIssuer,
					flagSignature,
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

func TestGetChecksum(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name          string
		checksumNames map[string]struct{}
		assetName     string
		want          *registry.Checksum
	}{
		{
			name: "sha512 gets priority over sha256",
			checksumNames: map[string]struct{}{
				"foo-1.0.0.tar.gz.sha256": {},
				"foo-1.0.0.tar.gz.sha512": {},
			},
			assetName: "foo-1.0.0.tar.gz",
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Asset:     "{{.Asset}}.sha512",
				Algorithm: sha512,
			},
		},
		{
			name: "sha512 gets priority over sha256 sha1 and md5",
			checksumNames: map[string]struct{}{
				"foo.tar.gz.md5":    {},
				"foo.tar.gz.sha1":   {},
				"foo.tar.gz.sha256": {},
				"foo.tar.gz.sha512": {},
			},
			assetName: "foo.tar.gz",
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Asset:     "{{.Asset}}.sha512",
				Algorithm: sha512,
			},
		},
		{
			name: "sha256 gets priority over sha1 and md5",
			checksumNames: map[string]struct{}{
				"foo.tar.gz.md5":    {},
				"foo.tar.gz.sha1":   {},
				"foo.tar.gz.sha256": {},
			},
			assetName: "foo.tar.gz",
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Asset:     "{{.Asset}}.sha256",
				Algorithm: "sha256",
			},
		},
		{
			name: "sha1 gets priority over md5",
			checksumNames: map[string]struct{}{
				"foo.tar.gz.md5":  {},
				"foo.tar.gz.sha1": {},
			},
			assetName: "foo.tar.gz",
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Asset:     "{{.Asset}}.sha1",
				Algorithm: "sha1",
			},
		},
		{
			name: "md5 is selected when only md5 is present",
			checksumNames: map[string]struct{}{
				"foo.tar.gz.md5": {},
			},
			assetName: "foo.tar.gz",
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Asset:     "{{.Asset}}.md5",
				Algorithm: "md5",
			},
		},
		{
			name: "no checksum returned when no matching checksum names exist",
			checksumNames: map[string]struct{}{
				"foo.tar.gz.asc": {},
			},
			assetName: "foo.tar.gz",
			want:      nil,
		},
		{
			name: "no checksum returned when checksum names is empty",
			checksumNames: map[string]struct{}{
				"bar.tar.gz.sha256": {},
			},
			assetName: "foo.tar.gz",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getChecksum(tt.checksumNames, tt.assetName)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getChecksum() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
