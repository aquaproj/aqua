package minisign_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestVerifier_Verify(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name             string
		isErr            bool
		fs               afero.Fs
		rt               *runtime.Runtime
		file             *download.File
		m                *registry.Minisign
		art              *template.Artifact
		param            *minisign.ParamVerify
		exe              minisign.Executor
		verifiedFilePath string
	}{
		{
			name: "normal",
			fs:   afero.NewMemMapFs(),
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
			m: &registry.Minisign{
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
			param: &minisign.ParamVerify{
				ArtifactPath: "/temp/foo",
			},
			exe: &minisign.MockExecutor{},
		},
	}
	ctx := context.Background()
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			verifier := minisign.New(d.fs, d.exe)
			if err := verifier.Verify(ctx, logE, d.param); err != nil {
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
