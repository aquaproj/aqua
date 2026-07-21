package initcmd_test

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
)

func TestController_Init(t *testing.T) {
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
				CWD:            "/home/foo/workspace",
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
				CWD:            "/home/foo/workspace",
				MaxParallelism: 5,
			},
			files: map[string]string{},
			releases: []*github.RepositoryRelease{
				{
					TagName: "v2.16.0",
				},
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
			}
			ctrl := initcmd.New(gh)
			// The configuration file is created relative to the working
			// directory when no path is given, so the test must give one.
			if err := ctrl.Init(ctx, logger, filepath.Join(dir, "aqua.yaml"), &initcmd.Param{}); err != nil {
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
