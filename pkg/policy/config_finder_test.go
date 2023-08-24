package policy_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/spf13/afero"
)

func TestConfigFinderImpl_Find(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		wd             string
		configFilePath string
		exp            string
		files          map[string]string
		dirs           map[string]struct{}
		isErr          bool
	}{
		{
			name: "not found",
			wd:   "/home/foo/bar",
			files: map[string]string{
				"/home/foo/.git": "",
			},
		},
		{
			name: ".git not found",
			wd:   "/home/foo",
		},
		{
			name:           "configFilePath",
			wd:             "/home/foo",
			configFilePath: "/home/foo/bar/aqua-policy.yaml",
			exp:            "/home/foo/bar/aqua-policy.yaml",
			files: map[string]string{
				"/home/foo/bar/aqua-policy.yaml": "",
			},
		},
		{
			name: "find",
			wd:   "/home/foo/bar",
			files: map[string]string{
				"/home/foo/aqua/aqua-policy.yaml": "",
			},
			dirs: map[string]struct{}{
				"/home/foo/.git": {},
			},
			exp: "/home/foo/aqua/aqua-policy.yaml",
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
			for name := range d.dirs {
				if err := osfile.MkdirAll(fs, name); err != nil {
					t.Fatal(err)
				}
			}
			configFinder := policy.NewConfigFinder(fs)
			p, err := configFinder.Find(d.configFilePath, d.wd)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returend")
			}
			if p != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, p)
			}
		})
	}
}
