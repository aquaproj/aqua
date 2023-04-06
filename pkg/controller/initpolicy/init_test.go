package initpolicy_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/initpolicy"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/sirupsen/logrus"
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
			files: map[string]string{
				"aqua-policy.yaml": `registries:
- type: standard
  ref: semver(">= 3.0.0")
packages:
- registry: standard
`,
			},
		},
		{
			name: "normal",
			param: &config.Param{
				PWD: "/home/foo/workspace",
			},
			files: map[string]string{},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			ctrl := initpolicy.New(fs)
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
