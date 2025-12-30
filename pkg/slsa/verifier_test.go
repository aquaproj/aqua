package slsa_test

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/spf13/afero"
)

func TestVerifier_Verify(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name             string
		isErr            bool
		fs               afero.Fs
		downloader       download.ClientAPI
		rt               *runtime.Runtime
		file             *download.File
		sp               *registry.SLSAProvenance
		art              *template.Artifact
		param            *slsa.ParamVerify
		exe              slsa.Executor
		verifiedFilePath string
	}{
		{
			name: "normal",
			fs:   afero.NewMemMapFs(),
			downloader: &download.Mock{
				RC: io.NopCloser(strings.NewReader("hello")),
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			file: &download.File{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Version:   "v1.6.0",
				Asset:     "aqua_darwin_arm64.tar.gz",
			},
			sp: &registry.SLSAProvenance{
				Type:      "github_release",
				RepoOwner: "aquaproj",
				RepoName:  "aqua",
				Asset:     ptr.String("multiple.intoto.jsonl"),
			},
			art: &template.Artifact{
				Version: "v1.6.0",
				OS:      "darwin",
				Arch:    "arm64",
				Format:  "tar.gz",
				Asset:   "aqua_darwin_arm64.tar.gz",
			},
			param: &slsa.ParamVerify{
				SourceURI:    "github.com/aquaproj/aqua",
				SourceTag:    "v1.6.0",
				ArtifactPath: "/temp/foo",
			},
			exe: &slsa.MockExecutor{},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			verifier := slsa.New(d.downloader, d.fs, d.exe)
			if err := verifier.Verify(ctx, logger, d.rt, d.sp, d.art, d.file, d.param); err != nil {
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
