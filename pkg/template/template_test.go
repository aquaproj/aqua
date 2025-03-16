package template_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/template"
)

func TestExecute(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		s     string
		input any
		exp   string
		isErr bool
	}{
		{
			name: "normal",
			s:    "foo-{{trimV .Version}}-{{.GOOS}}-{{.GOARCH}}.tar.gz",
			input: map[string]any{
				"Version": "v1.0.0",
				"GOOS":    "darwin",
				"GOARCH":  "amd64",
			},
			exp: "foo-1.0.0-darwin-amd64.tar.gz",
		},
		{
			name: "invalid template",
			s:    "foo-{{trimV .Version",
			input: map[string]any{
				"Version": "v1.0.0",
				"GOOS":    "darwin",
				"GOARCH":  "amd64",
			},
			isErr: true,
		},
		{
			name: "asset ext",
			s:    "{{trimAssetExt .Asset}}",
			input: map[string]string{
				"Asset": "foo-1.0.0-darwin-amd64.tar.gz",
			},
			exp: "foo-1.0.0-darwin-amd64",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			s, err := template.Execute(d.s, d.input)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if d.exp != s {
				t.Fatalf("wanted %s, got %s", d.exp, s)
			}
		})
	}
}
