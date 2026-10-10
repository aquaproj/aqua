package packslip_test

import (
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

func TestVerifier_Verify(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		ps    *registry.Packslip
		isErr bool
		exp   *packslip.ParamVerify
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
			verifier := packslip.New(&download.Mock{
				RC: io.NopCloser(strings.NewReader("bundle")),
			}, exe)
			err := verifier.Verify(t.Context(), logger, &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			}, d.ps, &template.Artifact{
				Version: "v2.5.0",
			}, &download.File{
				Type:      "github_release",
				RepoOwner: "jdx",
				RepoName:  "hk",
				Version:   "v2.5.0",
			}, artifactPath, "hk-aarch64-apple-darwin.tar.gz")
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			// packslip looks the artifact up by its file name, so it must be checked
			// under the name it is published under.
			if base := filepath.Base(exe.Param.ArtifactPath); base != "hk-aarch64-apple-darwin.tar.gz" {
				t.Fatalf("the artifact is checked as %s", base)
			}
			if base := filepath.Base(exe.Param.BundlePath); base != registry.BundleAssetPackslip {
				t.Fatalf("the bundle is %s", base)
			}
			if diff := cmp.Diff(d.exp, exe.Param, cmpopts.IgnoreFields(packslip.ParamVerify{}, "BundlePath", "ArtifactPath")); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
