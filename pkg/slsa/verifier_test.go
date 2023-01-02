package slsa_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/slsa"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func strP(s string) *string {
	return &s
}

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
				Asset:     strP("multiple.intoto.jsonl"),
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
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			verifier := slsa.New(d.downloader, d.fs, d.exe)
			if err := verifier.Verify(ctx, logE, d.rt, d.sp, d.art, d.file, d.param); err != nil {
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
