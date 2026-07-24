package cosign_test

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

func TestVerifier_Verify(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name             string
		isErr            bool
		executor         cosign.Executor
		downloader       download.ClientAPI
		cosignExePath    string
		rt               *runtime.Runtime
		param            *config.Param
		file             *download.File
		cos              *registry.Cosign
		art              *template.Artifact
		verifiedFilePath string
	}{
		{
			name:     "normal",
			executor: &osexec.Mock{},
			downloader: &download.Mock{
				RC: io.NopCloser(strings.NewReader("hello")),
			},
			cosignExePath: "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/sigstore/cosign/v1.13.1/cosign-darwin-arm64/cosign-darwin-arm64",
			rt: &runtime.Runtime{
				GOOS:   osDarwin,
				GOARCH: archArm64,
			},
			file: &download.File{
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  pkgNameAquaInstaller,
				Version:   versionV113,
				Path:      pkgNameAquaInstaller,
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			cos: &registry.Cosign{
				Opts: []string{
					"--signature",
					"https://github.com/aquaproj/aqua-installer/releases/download/{{.Version}}/aqua-installer.sig",
					"--certificate",
					"https://github.com/aquaproj/aqua-installer/releases/download/{{.Version}}/aqua-installer.pem",
				},
			},
			art: &template.Artifact{
				Version: versionV113,
				OS:      osDarwin,
				Arch:    archArm64,
				Format:  "raw",
				Asset:   pkgNameAquaInstaller,
			},
		},
		{
			name:     "signature, key, certificate",
			executor: &osexec.Mock{},
			downloader: &download.Mock{
				RC: io.NopCloser(strings.NewReader("hello")),
			},
			cosignExePath: "/home/foo/.local/share/aquaproj-aqua/pkgs/github_release/github.com/sigstore/cosign/v1.13.1/cosign-darwin-arm64/cosign-darwin-arm64",
			rt: &runtime.Runtime{
				GOOS:   osDarwin,
				GOARCH: archArm64,
			},
			file: &download.File{
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  pkgNameAquaInstaller,
				Version:   versionV113,
				Path:      pkgNameAquaInstaller,
			},
			param: &config.Param{
				RootDir: "/home/foo/.local/share/aquaproj-aqua",
			},
			cos: &registry.Cosign{
				Signature: &registry.DownloadedFile{
					Type:  pkgTypeGitHubRelease,
					Asset: new("aqua-installer.sig"),
				},
				Certificate: &registry.DownloadedFile{
					Type:  pkgTypeGitHubRelease,
					Asset: new("aqua-installer.pem"),
				},
				Key: &registry.DownloadedFile{
					Type:  pkgTypeGitHubRelease,
					Asset: new("aqua-installer.key"),
				},
			},
			art: &template.Artifact{
				Version: versionV113,
				OS:      osDarwin,
				Arch:    archArm64,
				Format:  "raw",
				Asset:   pkgNameAquaInstaller,
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			verifier := cosign.NewVerifier(d.executor, d.downloader, d.param)
			if err := verifier.Verify(ctx, logger, d.rt, d.file, d.cos, d.art, d.verifiedFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
