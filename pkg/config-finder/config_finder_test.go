package finder_test

import (
	"reflect"
	"testing"

	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/google/go-cmp/cmp"
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
			configFilePath: "/home/foo/aqua.yaml",
			exp:            "/home/foo/aqua.yaml",
		},
		{
			name: "find",
			wd:   "/home/foo",
			files: map[string]string{
				"/home/foo/.aqua.yaml": "",
			},
			exp: "/home/foo/.aqua.yaml",
		},
		{
			name: "global config",
			wd:   "/home/foo",
			globalConfigFilePaths: []string{
				"/home/.config/aqua.yaml",
				"/etc/aqua/aqua.yaml",
			},
			files: map[string]string{
				"/etc/aqua/aqua.yaml": "",
			},
			exp: "/etc/aqua/aqua.yaml",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
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

func Test_configFinderFinds(t *testing.T) {
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
			configFilePath: "/home/foo/aqua-2.yaml",
			exp:            []string{"/home/foo/aqua-2.yaml"},
		},
		{
			name: "find",
			wd:   "/home/foo",
			files: map[string]string{
				"/home/foo/.aqua.yaml": "",
				"/home/aqua.yaml":      "",
			},
			exp: []string{
				"/home/foo/.aqua.yaml",
				"/home/aqua.yaml",
			},
		},
		{
			name:           "find and configFilePath",
			wd:             "/home/foo",
			configFilePath: "aqua-2.yaml",
			files: map[string]string{
				"/home/.aqua.yaml": "",
			},
			exp: []string{
				"/home/foo/aqua-2.yaml",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs, err := testutil.NewFs(d.files)
			if err != nil {
				t.Fatal(err)
			}
			configFinder := finder.NewConfigFinder(fs)
			arr := configFinder.Finds(d.wd, d.configFilePath)
			if diff := cmp.Diff(d.exp, arr); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
