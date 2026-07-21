package finder_test

import (
	"path/filepath"
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
	pwd := "/home/foo"
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			paths := finder.ParseGlobalConfigFilePaths(pwd, d.env)
			if !reflect.DeepEqual(d.exp, paths) {
				t.Fatalf("wanted %+v, got %+v", d.exp, paths)
			}
		})
	}
}

// join makes the slash separated paths of a test case absolute.
func join(root string, paths ...string) []string {
	arr := make([]string, len(paths))
	for i, p := range paths {
		arr[i] = filepath.Join(root, filepath.FromSlash(p))
	}
	return arr
}

func Test_configFinderFind(t *testing.T) { //nolint:funlen
	t.Parallel()
	// Every path is relative to a root directory created for each test case,
	// except configFilePath when absConfigFilePath is false.
	data := []struct {
		name                  string
		wd                    string
		configFilePath        string
		absConfigFilePath     bool
		globalConfigFilePaths []string
		exp                   string
		files                 map[string]string
		isErr                 bool
	}{
		{
			name:  "not found",
			wd:    "foo",
			isErr: true,
		},
		{
			name:              "configFilePath",
			wd:                "foo",
			configFilePath:    "foo/aqua.yaml",
			absConfigFilePath: true,
			exp:               "foo/aqua.yaml",
		},
		{
			name: "find",
			wd:   "foo",
			files: map[string]string{
				"foo/.aqua.yaml": "",
			},
			exp: "foo/.aqua.yaml",
		},
		{
			name: "global config",
			wd:   "foo",
			globalConfigFilePaths: []string{
				"config/aqua.yaml",
				"etc/aqua/aqua.yaml",
			},
			files: map[string]string{
				"etc/aqua/aqua.yaml": "",
			},
			exp: "etc/aqua/aqua.yaml",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			testutil.WriteFiles(t, root, d.files, d.wd)
			configFilePath := d.configFilePath
			if d.absConfigFilePath {
				configFilePath = join(root, d.configFilePath)[0]
			}
			configFinder := finder.NewConfigFinder()
			p, err := configFinder.Find(join(root, d.wd)[0], configFilePath, join(root, d.globalConfigFilePaths...)...)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if exp := join(root, d.exp)[0]; p != exp {
				t.Fatalf("wanted %v, got %v", exp, p)
			}
		})
	}
}

func Test_configFinderFinds(t *testing.T) {
	t.Parallel()
	// Every path is relative to a root directory created for each test case,
	// except configFilePath, which is passed to Finds as it is.
	data := []struct {
		name           string
		wd             string
		configFilePath string
		files          map[string]string
		exp            []string
	}{
		{
			name: "not found",
			wd:   "home/foo",
		},
		{
			name: "find",
			wd:   "home/foo",
			files: map[string]string{
				"home/foo/.aqua.yaml": "",
				"home/aqua.yaml":      "",
			},
			exp: []string{
				"home/foo/.aqua.yaml",
				"home/aqua.yaml",
			},
		},
		{
			name:           "find and configFilePath",
			wd:             "home/foo",
			configFilePath: "aqua-2.yaml",
			files: map[string]string{
				"home/.aqua.yaml": "",
			},
			exp: []string{
				"home/foo/aqua-2.yaml",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			testutil.WriteFiles(t, root, d.files, d.wd)
			configFinder := finder.NewConfigFinder()
			arr := configFinder.Finds(join(root, d.wd)[0], d.configFilePath)
			var exp []string
			if d.exp != nil {
				exp = join(root, d.exp...)
			}
			if diff := cmp.Diff(exp, arr); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_DuplicateFilePaths(t *testing.T) {
	t.Parallel()
	cfgFilePaths := finder.ConfigFileNames()
	data := []struct {
		name     string
		filePath string
		exp      []string
	}{
		{
			name:     "not file",
			filePath: "yoo.yaml",
			exp:      nil,
		},
		{
			name:     "aqua.yaml",
			filePath: "aqua.yaml",
			exp:      cfgFilePaths,
		},
		{
			name:     "aqua/aqua.yaml",
			filePath: "aqua/aqua.yaml",
			exp:      cfgFilePaths,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			arr := finder.DuplicateFilePaths(d.filePath)
			m := make(map[string]struct{}, len(arr))
			for _, p := range arr {
				m[p] = struct{}{}
			}
			m2 := make(map[string]struct{}, len(d.exp))
			for _, p := range d.exp {
				m2[p] = struct{}{}
			}
			if diff := cmp.Diff(m2, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
