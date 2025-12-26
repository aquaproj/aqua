package list_test

import (
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/controller/list"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/sirupsen/logrus"
)

func TestController_List(t *testing.T) {
	t.Parallel()
	data := []struct {
		name              string
		files             map[string]string
		param             *config.Param
		registryInstaller list.RegistryInstaller
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
	downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logE, http.DefaultClient))
	rt := &runtime.Runtime{}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			httpDownloader := download.NewHTTPDownloader(logE, http.DefaultClient)
			ctrl := list.NewController(finder.NewConfigFinder(fs), reader.New(fs, d.param), registry.New(d.param, downloader, httpDownloader, fs, rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), fs)
			if err := ctrl.List(ctx, logE, d.param); err != nil {
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
