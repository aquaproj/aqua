package ast_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/controller/update/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestUpdateRegistries(t *testing.T) {
	t.Parallel()
	data := []struct {
		name        string
		updated     bool
		isErr       bool
		file        string
		newVersions map[string]string
		expFile     string
	}{
		{
			name: "updated",
			file: `registries:
  - type: standard
    ref: v4.0.0
    # foo
  - name: local
    type: local
    path: registry.yaml
  - name: custom
    type: github_content
    repo_owner: suzuki-shunsuke
    repo_name: aqua-registry
    ref: v3.0.0
    path: registry.yaml
`,
			expFile: `registries:
  - type: standard
    ref: v4.5.0
    # foo
  - name: local
    type: local
    path: registry.yaml
  - name: custom
    type: github_content
    repo_owner: suzuki-shunsuke
    repo_name: aqua-registry
    ref: v4.0.0
    path: registry.yaml
`,
			newVersions: map[string]string{
				"standard": "v4.5.0",
				"custom":   "v4.0.0",
			},
			updated: true,
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			file, err := parser.ParseBytes([]byte(d.file), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			updated, err := ast.UpdateRegistries(logE, file, d.newVersions)
			if err != nil {
				if !d.isErr {
					t.Fatal(err)
				}
			}
			if updated != d.updated {
				t.Fatalf("updated: wanted %v, got %v", d.updated, updated)
			}
			if diff := cmp.Diff(file.String(), d.expFile); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
