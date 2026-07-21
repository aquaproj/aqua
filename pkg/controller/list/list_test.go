package list_test

import (
	"log/slog"
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
				ConfigFilePath: "aqua.yaml",
				MaxParallelism: 5,
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
	rt := &runtime.Runtime{}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			d.param.CWD = t.TempDir()
			testutil.WriteFiles(t, d.param.CWD, d.files)
			ctrl := list.NewController(finder.NewConfigFinder(), reader.New(d.param), registry.New(d.param, downloader, rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}))
			if err := ctrl.List(ctx, logger, d.param); err != nil {
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
