package list_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	"github.com/clivm/clivm/pkg/controller/list"
	"github.com/clivm/clivm/pkg/download"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func TestController_List(t *testing.T) {
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		param             *config.Param
		registryInstaller registry.Installer
		isErr             bool
	}{
		{
			name: "normal",
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: clivm
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	downloader := download.NewRegistryDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
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
			ctrl := list.NewController(finder.NewConfigFinder(fs), reader.New(fs), registry.New(d.param, downloader, fs))
			if err := ctrl.List(ctx, d.param, logE); err != nil {
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
