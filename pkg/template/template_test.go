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
		input interface{}
		exp   string
		isErr bool
	}{
		{
			name: "normal",
			s:    "foo-{{trimV .Version}}-{{.GOOS}}-{{.GOARCH}}.tar.gz",
			input: map[string]interface{}{
				"Version": "v1.0.0",
				"GOOS":    "darwin",
				"GOARCH":  "amd64",
			},
			exp: "foo-1.0.0-darwin-amd64.tar.gz",
		},
		{
			name: "invalid template",
			s:    "foo-{{trimV .Version",
			input: map[string]interface{}{
				"Version": "v1.0.0",
				"GOOS":    "darwin",
				"GOARCH":  "amd64",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
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
