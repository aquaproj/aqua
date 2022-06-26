package finder_test

import (
	"reflect"
	"testing"

	finder "github.com/clivm/clivm/pkg/config-finder"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

func TestParseGlobalConfigFilePaths(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		env  string
		exp  []string
	}{
		{
			name: "empty",
			exp:  []string{},
		},
		{
			name: "normal",
			env:  ":/foo/bar:/yoo:/foo/bar",
			exp:  []string{"/foo/bar", "/yoo"},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			paths := finder.ParseGlobalConfigFilePaths(d.env)
			if !reflect.DeepEqual(d.exp, paths) {
				t.Fatalf("wanted %+v, got %+v", d.exp, paths)
			}
		})
	}
}

func Test_configFinderFind(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name                  string
		wd                    string
		configFilePath        string
		globalConfigFilePaths []string
		exp                   string
		files                 map[string]string
		isErr                 bool
	}{
		{
			name:  "not found",
			wd:    "/home/foo",
			isErr: true,
		},
		{
			name:           "configFilePath",
			wd:             "/home/foo",
			configFilePath: "/home/foo/clivm.yaml",
			exp:            "/home/foo/clivm.yaml",
		},
		{
			name: "find",
			wd:   "/home/foo",
			files: map[string]string{
				"/home/foo/.clivm.yaml": "",
			},
			exp: "/home/foo/.clivm.yaml",
		},
		{
			name: "global config",
			wd:   "/home/foo",
			globalConfigFilePaths: []string{
				"/home/.config/clivm.yaml",
				"/etc/aqua/clivm.yaml",
			},
			files: map[string]string{
				"/etc/aqua/clivm.yaml": "",
			},
			exp: "/etc/aqua/clivm.yaml",
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
			configFinder := finder.NewConfigFinder(fs)
			p, err := configFinder.Find(d.wd, d.configFilePath, d.globalConfigFilePaths...)
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

func Test_configFinderFinds(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		wd             string
		configFilePath string
		files          map[string]string
		exp            []string
	}{
		{
			name: "not found",
			wd:   "/home/foo",
			exp:  nil,
		},
		{
			name:           "configFilePath",
			wd:             "/home/foo",
			configFilePath: "/home/foo/clivm-2.yaml",
			exp:            []string{"/home/foo/clivm-2.yaml"},
		},
		{
			name: "find",
			wd:   "/home/foo",
			files: map[string]string{
				"/home/foo/.clivm.yaml": "",
				"/home/clivm.yaml":      "",
			},
			exp: []string{
				"/home/foo/.clivm.yaml",
				"/home/clivm.yaml",
			},
		},
		{
			name:           "find and configFilePath",
			wd:             "/home/foo",
			configFilePath: "clivm-2.yaml",
			files: map[string]string{
				"/home/.clivm.yaml": "",
			},
			exp: []string{
				"clivm-2.yaml",
				"/home/.clivm.yaml",
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
			configFinder := finder.NewConfigFinder(fs)
			arr := configFinder.Finds(d.wd, d.configFilePath)
			if diff := cmp.Diff(d.exp, arr); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
