package cosign

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

func TestHasFileOpt(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		opts []string
		exp  bool
	}{
		{
			name: "identity only",
			opts: []string{"--certificate-identity-regexp", "https://github.com/suzuki-shunsuke/go-release-workflow/.*", "--certificate-oidc-issuer", "https://token.actions.githubusercontent.com"},
		},
		{
			name: "signature",
			opts: []string{"--signature", "https://example.com/foo.sig"},
			exp:  true,
		},
		{
			name: "certificate with =",
			opts: []string{"--certificate=https://example.com/foo.pem"},
			exp:  true,
		},
		{
			name: "key",
			opts: []string{"--key", "https://example.com/cosign.pub"},
			exp:  true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if got := hasFileOpt(d.opts); got != d.exp {
				t.Fatalf("hasFileOpt = %v, wanted %v", got, d.exp)
			}
		})
	}
}

func TestIsNewBundle(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		content string
		exp     bool
	}{
		{
			name:    "new format",
			content: `{"mediaType":"application/vnd.dev.sigstore.bundle.v0.3+json","verificationMaterial":{}}`,
			exp:     true,
		},
		{
			name:    "old format",
			content: `{"base64Signature":"MEUCIQ...","cert":"LS0t","rekorBundle":{}}`,
		},
		{
			name:    "not JSON",
			content: "MEUCIQ...",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(t.TempDir(), "bundle")
			if err := os.WriteFile(p, []byte(d.content), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := isNewBundle(p)
			if err != nil {
				t.Fatal(err)
			}
			if got != d.exp {
				t.Fatalf("isNewBundle = %v, wanted %v", got, d.exp)
			}
		})
	}
}

// TestVerifier_VerifyDigest_unusable checks that a configuration cosign can't verify
// a digest with is turned down before anything is downloaded or run.
func TestVerifier_VerifyDigest_unusable(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		cos  *registry.Cosign
	}{
		{
			name: "no bundle",
			cos: &registry.Cosign{
				Opts: []string{"--certificate-identity-regexp", ".*"},
			},
		},
		{
			name: "signature and certificate",
			cos: &registry.Cosign{
				Signature:   &registry.DownloadedFile{Type: "github_release", Asset: new("foo.sig")},
				Certificate: &registry.DownloadedFile{Type: "github_release", Asset: new("foo.pem")},
			},
		},
		{
			name: "signature in opts",
			cos: &registry.Cosign{
				Bundle: &registry.DownloadedFile{Type: "github_release", Asset: new("foo.sigstore.json")},
				Opts:   []string{"--signature", "https://example.com/foo.sig"},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	rt := &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			// Neither a downloader nor an executor is given: using either would panic.
			v := &Verifier{}
			ok, err := v.VerifyDigest(t.Context(), logger, rt, &download.File{}, d.cos, &template.Artifact{}, "sha256:00")
			if err != nil {
				t.Fatal(err)
			}
			if ok {
				t.Fatal("the digest must not be used")
			}
		})
	}
}
