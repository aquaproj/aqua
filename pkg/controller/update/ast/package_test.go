package ast_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/controller/update/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestUpdatePackages(t *testing.T) { //nolint:funlen
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
			file: `packages:
  - name: cli/cli@v2.0.0  # comment
  # foo
  - name: suzuki-shunsuke/tfcmt@v4.1.0
  - import: foo.yaml
  - name: suzuki-shunsuke/tfcmt@v4.1.0
    registry: custom
  - name: suzuki-shunsuke/ci-info
    version: v3.0.0
`,
			expFile: `packages:
  - name: cli/cli@v2.1.0 # comment
  # foo
  - name: suzuki-shunsuke/tfcmt@v4.1.0
  - import: foo.yaml
  - name: suzuki-shunsuke/tfcmt@v4.6.0
    registry: custom
  - name: suzuki-shunsuke/ci-info
    version: v3.0.0
`,
			newVersions: map[string]string{
				"standard,cli/cli":                 "v2.1.0",
				"standard,suzuki-shunsuke/ci-info": "v4.0.0",
				"custom,suzuki-shunsuke/tfcmt":     "v4.6.0",
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
			updated, err := ast.UpdatePackages(logE, file, d.newVersions)
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
