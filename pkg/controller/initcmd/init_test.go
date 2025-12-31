package initcmd_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/ptr"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
)

func TestController_Init(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		files    map[string]string
		param    *config.Param
		releases []*github.RepositoryRelease
		isErr    bool
	}{
		{
			name: "file already exists",
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: standard
  ref: v2.15.0
packages:
`,
			},
		},
		{
			name: "normal",
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				MaxParallelism: 5,
			},
			files: map[string]string{},
			releases: []*github.RepositoryRelease{
				{
					TagName: ptr.String("v2.16.0"),
				},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
			}
			ctrl := initcmd.New(gh, fs)
			if err := ctrl.Init(ctx, logger, "", &initcmd.Param{}); err != nil {
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
