package policy_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/spf13/afero"
)

func TestValidator_Allow(t *testing.T) {
	t.Parallel()
	data := []struct {
		name           string
		rootDir        string
		configFilePath string
		files          map[string]string
		dirs           map[string]struct{}
		isErr          bool
	}{
		{
			name:           "normal",
			rootDir:        "/home/foo/.local/share/aquaproj-aqua",
			configFilePath: "/home/foo/workspace/aqua-policy.yaml",
			files: map[string]string{
				"/home/foo/workspace/aqua-policy.yaml": "",
			},
		},
		{
			name:    "warn file exists",
			rootDir: "/home/foo/.local/share/aquaproj-aqua",
			files: map[string]string{
				"/home/foo/workspace/aqua-policy.yaml":                                                     "",
				"/home/foo/.local/share/aquaproj-aqua/policy-warnings/home/foo/workspace/aqua-policy.yaml": "",
			},
		},
	}
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
			for name := range d.dirs {
				if err := util.MkdirAll(fs, name); err != nil {
					t.Fatal(err)
				}
			}
			validator := policy.NewValidator(&config.Param{
				RootDir: d.rootDir,
			}, fs)
			if err := validator.Allow(d.configFilePath); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returend")
			}
		})
	}
}
