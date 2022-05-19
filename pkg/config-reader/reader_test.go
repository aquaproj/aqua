package reader_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	reader "github.com/aquaproj/aqua/pkg/config-reader"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func Test_configReader_Read(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		exp            *config.Config
		isErr          bool
		files          map[string]string
		configFilePath string
	}{
		{
			name:  "file isn't found",
			isErr: true,
		},
		{
			name: "normal",
			files: map[string]string{
				"aqua.yaml": `registries:
- type: standard
  ref: v2.5.0
packages:`,
			},
			configFilePath: "aqua.yaml",
			exp: &config.Config{
				Registries: config.Registries{
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						Ref:       "v2.5.0",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
					},
				},
				Packages: []*config.Package{},
			},
		},
		{
			name: "import package",
			files: map[string]string{
				"aqua.yaml": `registries:
- type: standard
  ref: v2.5.0
packages:
- name: suzuki-shunsuke/ci-info@v1.0.0
- import: aqua-installer.yaml
`,
				"aqua-installer.yaml": `packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
			},
			configFilePath: "aqua.yaml",
			exp: &config.Config{
				Registries: config.Registries{
					"standard": {
						Type:      "github_content",
						Name:      "standard",
						Ref:       "v2.5.0",
						RepoOwner: "aquaproj",
						RepoName:  "aqua-registry",
						Path:      "registry.yaml",
					},
				},
				Packages: []*config.Package{
					{
						Name:     "suzuki-shunsuke/ci-info",
						Registry: "standard",
						Version:  "v1.0.0",
					},
					{
						Name:     "aquaproj/aqua-installer",
						Registry: "standard",
						Version:  "v1.0.0",
					},
				},
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
			reader := reader.New(fs)
			cfg := &config.Config{}
			if err := reader.Read(d.configFilePath, cfg); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if diff := cmp.Diff(d.exp, cfg); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
