package which_test

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func Test_controller_Which_caseInsensitive(t *testing.T) { //nolint:funlen
	t.Parallel()
	registryYAML := `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: upper
  path: upper
  files:
  - name: Foo
- type: github_content
  repo_owner: aquaproj
  repo_name: lower
  path: lower
  files:
  - name: foo
- type: github_content
  repo_owner: aquaproj
  repo_name: aliased
  path: aliased
  files:
  - name: bar
`
	data := []struct {
		name     string
		files    map[string]string
		goos     string
		global   bool
		exeName  string
		isErr    bool
		pkgName  string
		fileName string
	}{
		{
			name:    "case-insensitive match",
			exeName: "FOO",
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/lower@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: registryYAML,
			},
			pkgName:  "aquaproj/lower",
			fileName: "foo",
		},
		{
			name:    "exact match is preferred",
			exeName: "foo",
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/upper@v1.0.0
- name: aquaproj/lower@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: registryYAML,
			},
			pkgName:  "aquaproj/lower",
			fileName: "foo",
		},
		{
			name:    "exact match in a global config is preferred",
			exeName: "foo",
			global:  true,
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/upper@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: registryYAML,
				pathEtcAquaAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/lower@v1.0.0
`,
				pathEtcAquaRegistryYaml: registryYAML,
			},
			pkgName:  "aquaproj/lower",
			fileName: "foo",
		},
		{
			name:    "case-insensitive match of a command alias",
			exeName: "BAZ",
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aliased@v1.0.0
  command_aliases:
  - command: bar
    alias: baz
`,
				pathHomeFooWorkspaceRegistryYaml: registryYAML,
			},
			pkgName:  "aquaproj/aliased",
			fileName: "bar",
		},
		{
			name:    "upper-case extension on Windows",
			goos:    "windows",
			exeName: "foo.EXE",
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/lower@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: registryYAML,
			},
			pkgName:  "aquaproj/lower",
			fileName: "foo",
		},
		{
			name:    "the extension isn't trimmed on Linux",
			exeName: "foo.EXE",
			files: map[string]string{
				pathHomeFooWorkspaceAquaYaml: `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/lower@v1.0.0
`,
				pathHomeFooWorkspaceRegistryYaml: registryYAML,
			},
			isErr: true,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			goos := d.goos
			if goos == "" {
				goos = osLinux
			}
			rt := &runtime.Runtime{
				GOOS:   goos,
				GOARCH: archAmd64,
			}
			param := &config.Param{
				CWD:            pathHomeFooWorkspace,
				ConfigFilePath: "aqua.yaml",
				RootDir:        pathHomeFooLocalShare,
				MaxParallelism: 5,
			}
			if d.global {
				param.GlobalConfigFilePaths = []string{pathEtcAquaAquaYaml}
			}
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			testutil.RootParam(dir, param)
			env := testutil.RootEnv(dir, map[string]string{"PATH": ""})
			downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			ctrl := which.New(param, finder.NewConfigFinder(), reader.New(param), registry.New(param, downloader, rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), rt, osenv.NewMock(env), link.New())
			findResult, err := ctrl.Which(t.Context(), logger, param, d.exeName)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if findResult.Package == nil {
				t.Fatalf("a package must be found: %+v", findResult)
			}
			if name := findResult.Package.Package.Name; name != d.pkgName {
				t.Fatalf("package name: wanted %s, got %s", d.pkgName, name)
			}
			if name := findResult.File.Name; name != d.fileName {
				t.Fatalf("file name: wanted %s, got %s", d.fileName, name)
			}
		})
	}
}
