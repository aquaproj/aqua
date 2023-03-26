package initcmd_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/initcmd"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

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
					TagName: stringP("v2.16.0"),
				},
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			gh := &github.MockRepositoriesService{
				Releases: d.releases,
			}
			ctrl := initcmd.New(gh, fs)
			if err := ctrl.Init(ctx, "", logE); err != nil {
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
