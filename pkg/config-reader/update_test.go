package reader_test

import (
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func Test_configReader_ReadToUpdate(t *testing.T) { //nolint:funlen
	t.Parallel()
	// The keys of files and configFilePath are relative to a directory created
	// for each test case, which is also passed to cfg and cfgs so that they can
	// build the absolute paths the reader returns.
	data := []struct {
		name           string
		cfg            func(dir string) *aqua.Config
		cfgs           func(dir string) map[string]*aqua.Config
		isErr          bool
		files          map[string]string
		configFilePath string
		homeDir        string
	}{
		{
			name:           "file isn't found",
			configFilePath: fileAquaYaml,
			isErr:          true,
		},
		{
			name: "normal",
			files: map[string]string{
				fileAquaYaml: `registries:
- type: standard
  ref: v2.5.0
- type: local
  name: local
  path: registry.yaml
packages:`,
			},
			configFilePath: fileAquaYaml,
			cfg: func(dir string) *aqua.Config {
				return &aqua.Config{
					Registries: aqua.Registries{
						regTypeStandard: {
							Type:      pkgTypeGitHubContent,
							Name:      regTypeStandard,
							Ref:       "v2.5.0",
							RepoOwner: regOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Path:      regFileRegistryYaml,
						},
						regTypeLocal: {
							Type: regTypeLocal,
							Name: regTypeLocal,
							Path: filepath.Join(dir, regFileRegistryYaml),
						},
					},
					Packages: []*aqua.Package{},
				}
			},
			cfgs: func(string) map[string]*aqua.Config {
				return map[string]*aqua.Config{}
			},
		},
		{
			name: "import package",
			files: map[string]string{
				fileAquaYaml: `registries:
- type: standard
  ref: v2.5.0
packages:
- name: suzuki-shunsuke/ci-info@v1.0.0
- import: aqua-installer.yaml
`,
				fileAquaInstallerYaml: `packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
			},
			configFilePath: fileAquaYaml,
			cfg: func(string) *aqua.Config {
				return &aqua.Config{
					Registries: aqua.Registries{
						regTypeStandard: {
							Type:      pkgTypeGitHubContent,
							Name:      regTypeStandard,
							Ref:       "v2.5.0",
							RepoOwner: regOwnerAquaproj,
							RepoName:  regNameAquaRegistry,
							Path:      regFileRegistryYaml,
						},
					},
					Packages: []*aqua.Package{
						{
							Name:     "suzuki-shunsuke/ci-info",
							Registry: regTypeStandard,
							Version:  versionV1,
						},
					},
				}
			},
			cfgs: func(dir string) map[string]*aqua.Config {
				return map[string]*aqua.Config{
					filepath.Join(dir, fileAquaInstallerYaml): {
						Packages: []*aqua.Package{
							{
								Name:     "aquaproj/aqua-installer",
								Registry: regTypeStandard,
								Version:  versionV1,
							},
						},
						Registries: aqua.Registries{
							regTypeStandard: {
								Type:      pkgTypeGitHubContent,
								Name:      regTypeStandard,
								Ref:       "v2.5.0",
								RepoOwner: regOwnerAquaproj,
								RepoName:  regNameAquaRegistry,
								Path:      regFileRegistryYaml,
							},
						},
					},
				}
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			reader := reader.New(&config.Param{
				HomeDir: d.homeDir,
			})
			cfg := &aqua.Config{}
			cfgs, err := reader.ReadToUpdate(filepath.Join(dir, d.configFilePath), cfg)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.cfg(dir), cfg); diff != "" {
				t.Fatal("cfg:", diff)
			}
			if diff := cmp.Diff(d.cfgs(dir), cfgs); diff != "" {
				t.Fatal("cfgs:", diff)
			}
		})
	}
}
