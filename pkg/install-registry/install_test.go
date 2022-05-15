package registry_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func stringP(s string) *string {
	return &s
}

func Test_installer_InstallRegistries(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name        string
		files       map[string]string
		param       *config.Param
		downloader  download.RegistryDownloader
		cfg         *config.Config
		cfgFilePath string
		isErr       bool
		exp         map[string]*config.RegistryContent
	}{
		{
			name: "local",
			param: &config.Param{
				MaxParallelism: 5,
			},
			cfgFilePath: "aqua.yaml",
			files: map[string]string{
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
			cfg: &config.Config{
				Registries: config.Registries{
					"local": {
						Type: "local",
						Name: "local",
						Path: "registry.yaml",
					},
				},
			},
			exp: map[string]*config.RegistryContent{
				"local": {
					PackageInfos: config.PackageInfos{
						{
							Type:      "github_content",
							RepoOwner: "aquaproj",
							RepoName:  "aqua-installer",
							Path:      stringP("aqua-installer"),
						},
					},
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
			inst := registry.New(d.param, d.downloader, fs)
			registries, err := inst.InstallRegistries(ctx, d.cfg, d.cfgFilePath, logE)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, registries); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
