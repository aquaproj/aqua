package checksumgetter

import (
	"errors"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type fakeReleaseAssets map[string]*domain.AssetDigest

func (r fakeReleaseAssets) GetDigest(assetName string) *domain.AssetDigest {
	return r[assetName]
}

func TestNeedsAssetVerification(t *testing.T) {
	t.Parallel()
	signedChecksum := &registry.Checksum{
		Type:   "github_release",
		Asset:  "checksums.txt",
		Cosign: &registry.Cosign{Opts: []string{"--certificate-identity-regexp", ".*"}},
	}
	data := []struct {
		name    string
		pkgInfo *registry.PackageInfo
		exp     bool
	}{
		{
			name:    "no signature",
			pkgInfo: &registry.PackageInfo{},
		},
		{
			name:    "asset signature",
			pkgInfo: &registry.PackageInfo{GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{}},
			exp:     true,
		},
		{
			// The signed checksum file ties the digest to the publisher already.
			name:    "asset signature and a signed checksum file",
			pkgInfo: &registry.PackageInfo{GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{}, Checksum: signedChecksum},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := needsAssetVerification(d.pkgInfo); got != d.exp {
				t.Fatalf("needsAssetVerification = %v, wanted %v", got, d.exp)
			}
		})
	}
}

func TestGetter_verifyAndGetChecksum(t *testing.T) { //nolint:funlen
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	const (
		asset  = "ar2_linux_amd64.tar.gz"
		digest = "CD76FBE5150C06F4928C8C00759BF54237C85C4CFF829D4A6A19628709CA2AA7"
		// sha256 of "hello"
		helloSum = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	)
	data := []struct {
		name          string
		releaseAssets domain.ReleaseAssets
		verifier      *MockChecksumFileVerifier
		downloadErr   error
		exp           *checksum.Checksum
		expDigests    []string
		isErr         bool
	}{
		{
			// cosign verified the digest, so the asset isn't downloaded.
			name:          "verified with the digest",
			releaseAssets: fakeReleaseAssets{asset: {Digest: digest, Algorithm: "sha256"}},
			verifier:      &MockChecksumFileVerifier{DigestOK: true, Kinds: []string{"cosign"}},
			downloadErr:   errors.New("the asset must not be downloaded"),
			exp:           &checksum.Checksum{Checksum: digest, Algorithm: "sha256", Verified: []string{"cosign"}},
			expDigests:    []string{"sha256:" + strings.ToLower(digest)},
		},
		{
			// e.g. GitHub Artifact Attestations, which need the asset.
			name:          "the digest can't be used",
			releaseAssets: fakeReleaseAssets{asset: {Digest: digest, Algorithm: "sha256"}},
			verifier:      &MockChecksumFileVerifier{Kinds: []string{"github_artifact_attestations"}},
			exp:           &checksum.Checksum{Checksum: helloSum, Algorithm: "sha256", Verified: []string{"github_artifact_attestations"}},
			expDigests:    []string{"sha256:" + strings.ToLower(digest)},
		},
		{
			name:     "no digest",
			verifier: &MockChecksumFileVerifier{Kinds: []string{"slsa_provenance"}},
			exp:      &checksum.Checksum{Checksum: helloSum, Algorithm: "sha256", Verified: []string{"slsa_provenance"}},
		},
		{
			name:     "verification fails",
			verifier: &MockChecksumFileVerifier{Err: errors.New("invalid signature")},
			isErr:    true,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			pkg, cid := newGitHubReleasePkg(t, "aquaproj", "ar2", "v0.15.14", "ar2_{{.OS}}_{{.Arch}}.tar.gz", rt)
			pkg.PackageInfo.Cosign = &registry.Cosign{
				Bundle: &registry.DownloadedFile{Type: "github_release", Asset: new("ar2_{{.OS}}_{{.Arch}}.tar.gz.sigstore.json")},
			}
			var rc io.ReadCloser
			if d.downloadErr == nil {
				rc = io.NopCloser(strings.NewReader("hello"))
			}
			g := &Getter{
				downloader:           &download.Mock{RC: rc, Err: d.downloadErr},
				checksumFileVerifier: d.verifier,
			}
			cs, err := g.verifyAndGetChecksum(t.Context(), logger, pkg, rt, cid, d.releaseAssets)
			if d.isErr {
				if err == nil {
					t.Fatal("error must be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			d.exp.ID = cid
			assertChecksum(t, cs, d.exp)
			if !slices.Equal(d.verifier.Digests, d.expDigests) {
				t.Fatalf("digests = %v, wanted %v", d.verifier.Digests, d.expDigests)
			}
		})
	}
}

func assertChecksum(t *testing.T, got, exp *checksum.Checksum) {
	t.Helper()
	if got.ID != exp.ID || !strings.EqualFold(got.Checksum, exp.Checksum) || got.Algorithm != exp.Algorithm || !slices.Equal(got.Verified, exp.Verified) {
		t.Fatalf("checksum = %+v, wanted %+v", got, exp)
	}
}

func TestGetter_dlAndVerifyChecksumFile_verified(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	pkg, cid := newGitHubReleasePkg(t, "aquaproj", "ar2", "v0.15.14", "ar2_{{.OS}}_{{.Arch}}.tar.gz", rt)
	pkg.PackageInfo.Checksum = &registry.Checksum{
		Type:                       "github_release",
		Asset:                      "ar2_checksums.txt",
		Algorithm:                  "sha256",
		GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{},
	}
	g := &Getter{
		chkDL: &download.MockChecksumDownloader{
			Body: `2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824  ar2_linux_amd64.tar.gz
cd76fbe5150c06f4928c8c00759bf54237c85c4cff829d4a6a19628709ca2aa7  ar2_darwin_arm64.tar.gz
`,
		},
		checksumFileVerifier: &MockChecksumFileVerifier{ChecksumKinds: []string{"checksum.github_artifact_attestations"}},
	}
	cs, err := g.dlAndVerifyChecksumFile(t.Context(), slog.New(slog.DiscardHandler), pkg, rt,
		map[string]struct{}{"ar2_linux_amd64.tar.gz": {}, "ar2_darwin_arm64.tar.gz": {}}, cid)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 2 {
		t.Fatalf("wanted two checksums, got %d", len(cs))
	}
	// Every checksum the signed file lists records the file's verification.
	for _, c := range cs {
		if !slices.Equal(c.Verified, []string{"checksum.github_artifact_attestations"}) {
			t.Fatalf("checksum = %+v", c)
		}
	}
}
