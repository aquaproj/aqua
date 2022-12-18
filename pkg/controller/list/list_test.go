package list_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/aquaproj/aqua/pkg/controller/list"
	"github.com/aquaproj/aqua/pkg/cosign"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/download"
	registry "github.com/aquaproj/aqua/pkg/install-registry"
	"github.com/aquaproj/aqua/pkg/runtime"
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
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
	rt := &runtime.Runtime{}
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
			ctrl := list.NewController(finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, downloader, fs, rt, &cosign.MockVerifier{}), &domain.MockCosignInstaller{})
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
