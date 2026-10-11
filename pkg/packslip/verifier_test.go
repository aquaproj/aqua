package packslip_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/packslip"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// recordingDownloader records which file is downloaded.
type recordingDownloader struct {
	file *download.File
}

func (r *recordingDownloader) ReadCloser(_ context.Context, _ *slog.Logger, file *download.File) (io.ReadCloser, int64, error) {
	r.file = file
	return io.NopCloser(strings.NewReader("bundle")), 0, nil
}

func TestVerifier_Verify(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		pkgType string
		ps      *registry.Packslip
		isErr   bool
		exp     *packslip.ParamVerify
		// expFile is where the bundle is downloaded from.
		expFile *download.File
	}{
		{
			name: "identity",
			ps: &registry.Packslip{
				IdentityPrefix: "https://github.com/jdx/hk/",
				Issuer:         "https://token.actions.githubusercontent.com",
				AllowUnlogged:  true,
			},
			exp: &packslip.ParamVerify{
				IdentityPrefix: "https://github.com/jdx/hk/",
				Issuer:         "https://token.actions.githubusercontent.com",
				AllowUnlogged:  true,
			},
			expFile: &download.File{
				Type:      "github_release",
				RepoOwner: "jdx",
				RepoName:  "hk",
				Version:   "v2.5.0",
				Asset:     registry.BundleAssetPackslip,
			},
		},
		{
			name: "pin",
			ps: &registry.Packslip{
				Pin: "ps1_snirenkjwr7m5ozgcufameodnm",
			},
			exp: &packslip.ParamVerify{
				Pin: "ps1_snirenkjwr7m5ozgcufameodnm",
			},
		},
		{
			name:    "url",
			pkgType: "http",
			ps: &registry.Packslip{
				URL:            new("https://example.com/hk/{{.Version}}/packslip.sigstore.json"),
				IdentityPrefix: "https://github.com/jdx/hk/",
			},
			exp: &packslip.ParamVerify{
				IdentityPrefix: "https://github.com/jdx/hk/",
			},
			expFile: &download.File{
				Type:    "http",
				Version: "v2.5.0",
				URL:     "https://example.com/hk/v2.5.0/packslip.sigstore.json",
			},
		},
		{
			// A package that isn't a release asset has no release to find the bundle in.
			name:    "http package without url",
			pkgType: "http",
			ps: &registry.Packslip{
				IdentityPrefix: "https://github.com/jdx/hk/",
			},
			isErr: true,
		},
		{
			// Without whom to trust, packslip would trust whichever project the bundle claims.
			name: "no trust",
			ps: &registry.Packslip{
				Issuer: "https://token.actions.githubusercontent.com",
			},
			isErr: true,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			artifactPath := filepath.Join(t.TempDir(), "downloaded")
			if err := os.WriteFile(artifactPath, []byte("artifact"), 0o600); err != nil {
				t.Fatal(err)
			}
			exe := &packslip.MockExecutor{}
			downloader := &recordingDownloader{}
			verifier := packslip.New(downloader, exe)
			pkgType := d.pkgType
			if pkgType == "" {
				pkgType = "github_release"
			}
			file := &download.File{
				Type:    pkgType,
				Version: "v2.5.0",
			}
			if pkgType == "github_release" {
				file.RepoOwner = "jdx"
				file.RepoName = "hk"
			}
			err := verifier.Verify(t.Context(), logger, &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			}, d.ps, &template.Artifact{
				Version: "v2.5.0",
			}, file, artifactPath, "hk-aarch64-apple-darwin.tar.gz")
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if d.expFile != nil {
				if diff := cmp.Diff(d.expFile, downloader.file); diff != "" {
					t.Fatal(diff)
				}
			}
			checkParam(t, d.exp, exe.Param)
		})
	}
}

func checkParam(t *testing.T, exp, param *packslip.ParamVerify) {
	t.Helper()
	// packslip looks the artifact up by its file name, so it must be checked
	// under the name it is published under.
	if base := filepath.Base(param.ArtifactPath); base != "hk-aarch64-apple-darwin.tar.gz" {
		t.Fatalf("the artifact is checked as %s", base)
	}
	if base := filepath.Base(param.BundlePath); base != registry.BundleAssetPackslip {
		t.Fatalf("the bundle is %s", base)
	}
	if diff := cmp.Diff(exp, param, cmpopts.IgnoreFields(packslip.ParamVerify{}, "BundlePath", "ArtifactPath")); diff != "" {
		t.Fatal(diff)
	}
}
