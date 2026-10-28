package installpackage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type fakeVerifier struct {
	kind string
	err  error
}

func (v *fakeVerifier) Kind() string {
	return v.kind
}

func (v *fakeVerifier) Enabled(_ *slog.Logger) (bool, error) {
	return true, nil
}

func (v *fakeVerifier) Verify(_ context.Context, _ *slog.Logger, _ string) error {
	return v.err
}

func newVerifiedTestPackage() *config.Package {
	return &config.Package{
		Package: &aqua.Package{
			Name:    "aquaproj/ar2",
			Version: "v0.15.14",
		},
		PackageInfo: &registry.PackageInfo{
			Type:      "github_release",
			RepoOwner: "aquaproj",
			RepoName:  "ar2",
			Asset:     "ar2_{{.OS}}_{{.Arch}}.tar.gz",
		},
	}
}

func TestInstaller_skipRecordedVerifiers(t *testing.T) { //nolint:funlen
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	pkg := newVerifiedTestPackage()
	cid, err := pkg.ChecksumID(rt)
	if err != nil {
		t.Fatal(err)
	}
	verifiers := []FileVerifier{
		&fakeVerifier{kind: "github_artifact_attestations"},
		&fakeVerifier{kind: "cosign"},
	}
	data := []struct {
		name             string
		entry            *checksum.Checksum
		checksum         *checksum.Checksum
		signedChecksum   bool
		verifySignatures bool
		exp              int
	}{
		{
			// The checksum file's signature ties the digest to the publisher, as the
			// lock file relies on too.
			name:           "the signed checksum file recorded",
			signedChecksum: true,
			entry:          &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256", Verified: []string{"checksum.github_artifact_attestations"}},
		},
		{
			name:           "the signed checksum file not recorded",
			signedChecksum: true,
			entry:          &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256"},
			exp:            2,
		},
		{
			// The entry was recorded from a signed checksum file the registry no
			// longer configures.
			name:  "a checksum file kind the registry doesn't configure",
			entry: &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256", Verified: []string{"checksum.github_artifact_attestations"}},
			exp:   2,
		},
		{
			name: "no entry",
			exp:  2,
		},
		{
			// An entry written before the verification was recorded keeps its meaning.
			name:  "entry without the record",
			entry: &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256"},
			exp:   2,
		},
		{
			name:  "every kind recorded",
			entry: &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256", Verified: []string{"cosign", "github_artifact_attestations"}},
		},
		{
			// The registry asks for a kind the entry doesn't record, e.g. because it
			// added cosign after the entry was written.
			name:  "a kind missing",
			entry: &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256", Verified: []string{"github_artifact_attestations"}},
			exp:   2,
		},
		{
			name:             "asked to verify anyway",
			entry:            &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256", Verified: []string{"cosign", "github_artifact_attestations"}},
			verifySignatures: true,
			exp:              2,
		},
		{
			// A checksum given by the caller, e.g. from the lock file, isn't an
			// aqua-checksums.json entry.
			name:     "checksum given by the caller",
			checksum: &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256"},
			entry:    &checksum.Checksum{ID: cid, Checksum: "foo", Algorithm: "sha256", Verified: []string{"cosign", "github_artifact_attestations"}},
			exp:      2,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksums := checksum.New()
			if d.entry != nil {
				checksums.Set(cid, d.entry)
			}
			is := &Installer{
				runtime:          rt,
				verifySignatures: d.verifySignatures,
			}
			pkg := newVerifiedTestPackage()
			if d.signedChecksum {
				pkg.PackageInfo.Checksum = &registry.Checksum{
					Type:                       "github_release",
					Asset:                      "ar2_checksums.txt",
					Algorithm:                  "sha256",
					GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{},
				}
			}
			got, err := is.skipRecordedVerifiers(logger, &DownloadParam{
				Package:   pkg,
				Checksums: checksums,
				Checksum:  d.checksum,
				Asset:     "ar2_linux_amd64.tar.gz",
			}, verifiers)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != d.exp {
				t.Fatalf("%d verifiers are left, wanted %d", len(got), d.exp)
			}
		})
	}
}

func TestInstaller_runVerifiers(t *testing.T) {
	t.Parallel()
	is := &Installer{}
	logger := slog.New(slog.DiscardHandler)
	called := false
	path := func() (string, error) {
		called = true
		return "asset", nil
	}
	kinds, err := is.runVerifiers(t.Context(), logger, nil, path)
	if err != nil {
		t.Fatal(err)
	}
	if kinds != nil || called {
		t.Fatalf("nothing should be verified: kinds=%v, path called=%v", kinds, called)
	}
	kinds, err = is.runVerifiers(t.Context(), logger, []FileVerifier{
		&fakeVerifier{kind: "github_artifact_attestations"},
		&fakeVerifier{kind: "cosign"},
	}, path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(kinds, []string{"github_artifact_attestations", "cosign"}) {
		t.Fatalf("kinds = %v", kinds)
	}
	// gh couldn't authenticate: the install goes on, but the attestations weren't
	// verified, so they must not be recorded.
	kinds, err = is.runVerifiers(t.Context(), logger, []FileVerifier{
		&fakeVerifier{kind: "github_artifact_attestations", err: fmt.Errorf("verify a package with gh attestation: %w", ghattestation.ErrSkipped)},
		&fakeVerifier{kind: "cosign"},
	}, path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(kinds, []string{"cosign"}) {
		t.Fatalf("kinds = %v", kinds)
	}
	if _, err := is.runVerifiers(t.Context(), logger, []FileVerifier{
		&fakeVerifier{kind: "cosign", err: errors.New("invalid signature")},
	}, path); err == nil {
		t.Fatal("an invalid signature must fail")
	}
}

func TestInstaller_verifyChecksumFile_skipped(t *testing.T) {
	t.Parallel()
	is := &Installer{}
	verified, err := is.verifyChecksumFile(t.Context(), slog.New(slog.DiscardHandler), []byte("checksums"), []FileVerifier{
		&fakeVerifier{kind: "github_artifact_attestations", err: ghattestation.ErrSkipped},
		&fakeVerifier{kind: "cosign"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if kinds := checksumFileKinds(verified); !slices.Equal(kinds, []string{"checksum.cosign"}) {
		t.Fatalf("kinds = %v", kinds)
	}
}

func TestHasChecksumFileRecord(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		entry *checksum.Checksum
		exp   bool
	}{
		{name: "no entry"},
		{name: "no record", entry: &checksum.Checksum{}},
		{name: "the asset's own", entry: &checksum.Checksum{Verified: []string{"cosign"}}},
		{name: "the checksum file's", entry: &checksum.Checksum{Verified: []string{"cosign", "checksum.cosign"}}, exp: true},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := hasChecksumFileRecord(d.entry); got != d.exp {
				t.Fatalf("hasChecksumFileRecord = %v, wanted %v", got, d.exp)
			}
		})
	}
}

func TestInstaller_verifyChecksum_verified(t *testing.T) { //nolint:funlen
	t.Parallel()
	const (
		cid = "github_release/github.com/aquaproj/ar2/v0.15.14/ar2_linux_amd64.tar.gz"
		sum = "3516A4D84F7B69EA5752CA2416895A2705910AF3ED6815502AF789000FC7E963"
	)
	data := []struct {
		name  string
		entry *checksum.Checksum
		calc  string
		isErr bool
		exp   *checksum.Checksum
	}{
		{
			name: "a new entry records the kinds verified",
			calc: sum,
			exp:  &checksum.Checksum{ID: cid, Checksum: sum, Algorithm: "sha256", Verified: []string{"cosign"}},
		},
		{
			// The signatures were verified against this asset, which isn't the one the
			// entry describes.
			name:  "an existing entry is left as it is",
			entry: &checksum.Checksum{ID: cid, Checksum: sum, Algorithm: "sha256"},
			calc:  sum,
			exp:   &checksum.Checksum{ID: cid, Checksum: sum, Algorithm: "sha256"},
		},
		{
			name:  "a mismatch",
			entry: &checksum.Checksum{ID: cid, Checksum: sum, Algorithm: "sha256"},
			calc:  "0000000000000000000000000000000000000000000000000000000000000000",
			isErr: true,
			exp:   &checksum.Checksum{ID: cid, Checksum: sum, Algorithm: "sha256"},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksums := checksum.New()
			if d.entry != nil {
				checksums.Set(cid, d.entry)
			}
			is := &Installer{
				checksumCalculator: &MockChecksumCalculator{Checksum: d.calc},
			}
			err := is.verifyChecksum(t.Context(), logger, &ParamVerifyChecksum{
				ChecksumID: cid,
				Checksum:   checksums.Get(cid),
				Checksums:  checksums,
				Pkg:        newVerifiedTestPackage(),
				Verified:   []string{"cosign"},
			})
			if d.isErr {
				if err == nil {
					t.Fatal("error must be returned")
				}
			} else if err != nil {
				t.Fatal(err)
			}
			got := checksums.Get(cid)
			if got.Checksum != d.exp.Checksum || !slices.Equal(got.Verified, d.exp.Verified) {
				t.Fatalf("entry = %+v, wanted %+v", got, d.exp)
			}
		})
	}
}

func TestRecordChecksum(t *testing.T) {
	t.Parallel()
	const cid = "github_release/github.com/aquaproj/ar2/v0.15.14/ar2_linux_amd64.tar.gz"
	checksums := checksum.New()
	// A checksum taken from a signed checksum file carries the file's kinds, and the
	// asset's own are added to them.
	recordChecksum(checksums, cid, &checksum.Checksum{
		ID:        cid,
		Checksum:  "foo",
		Algorithm: "sha256",
		Verified:  []string{"checksum.cosign"},
	}, false, []string{"github_artifact_attestations"})
	got := checksums.Get(cid)
	if !slices.Equal(got.Verified, []string{"github_artifact_attestations", "checksum.cosign"}) {
		t.Fatalf("verified = %v", got.Verified)
	}
}

func TestInstaller_VerifyAssetDigest(t *testing.T) {
	t.Parallel()
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	data := []struct {
		name    string
		pkgInfo func(*registry.PackageInfo)
		exp     bool
	}{
		{
			// Nothing to verify, so the digest can be used as it is.
			name:    "no signature",
			pkgInfo: func(*registry.PackageInfo) {},
			exp:     true,
		},
		{
			// gh attestation verify has no digest input.
			name: "github artifact attestations",
			pkgInfo: func(p *registry.PackageInfo) {
				p.GitHubArtifactAttestations = &registry.GitHubArtifactAttestations{}
			},
		},
		{
			name: "slsa provenance",
			pkgInfo: func(p *registry.PackageInfo) {
				p.SLSAProvenance = &registry.SLSAProvenance{Type: "github_release"}
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			pkg := newVerifiedTestPackage()
			d.pkgInfo(pkg.PackageInfo)
			is := &Installer{runtime: rt, realRuntime: rt}
			kinds, ok, err := is.VerifyAssetDigest(t.Context(), logger, pkg, "ar2_linux_amd64.tar.gz", "sha256:00", rt)
			if err != nil {
				t.Fatal(err)
			}
			if ok != d.exp {
				t.Fatalf("ok = %v, wanted %v", ok, d.exp)
			}
			if len(kinds) != 0 {
				t.Fatalf("kinds = %v, wanted none", kinds)
			}
		})
	}
}
