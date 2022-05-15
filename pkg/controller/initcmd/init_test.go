package initcmd_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller/initcmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestController_Init(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		files map[string]string
		param *config.Param
		isErr bool
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
			ctrl := initcmd.New(nil, fs)
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
