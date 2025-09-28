package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
	"github.com/goccy/go-yaml/parser"
	"github.com/google/go-cmp/cmp"
)

func TestUpdateASTFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name     string
		yamlStr  string
		pkgs     any
		expected string
		wantErr  bool
	}{
		{
			name: "add packages to null",
			yamlStr: `packages: null
registries:
  - ref: v1.0.0
    type: standard`,
			pkgs: []map[string]string{
				{"name": "cli/cli"},
				{"name": "kubernetes/kubectl"},
			},
			expected: `packages:
- name: cli/cli
- name: kubernetes/kubectl
registries:
  - ref: v1.0.0
    type: standard
`,
			wantErr: false,
		},
		{
			name: "merge packages to existing array",
			yamlStr: `packages:
  - name: existing/package
registries:
  - ref: v1.0.0
    type: standard`,
			pkgs: []map[string]string{
				{"name": "cli/cli"},
			},
			expected: `packages:
  - name: existing/package
  - name: cli/cli
registries:
  - ref: v1.0.0
    type: standard
`,
			wantErr: false,
		},
		{
			name: "invalid packages type",
			yamlStr: `packages: "invalid"
registries:
  - ref: v1.0.0
    type: standard`,
			pkgs:    []map[string]string{{"name": "cli/cli"}},
			wantErr: true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			file, err := parser.ParseBytes([]byte(d.yamlStr), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}

			err = asset.UpdateASTFile(file, d.pkgs)
			if d.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if !d.wantErr {
				result := file.String()
				if diff := cmp.Diff(d.expected, result); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
