package policy_test

import (
	"maps"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
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
			configFilePath: pathHomeFooBarPolicy,
			exp:            pathHomeFooBarPolicy,
			files: map[string]string{
				pathHomeFooBarPolicy: "",
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
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			// The paths of the test cases are rooted at a temporary directory.
			dir := t.TempDir()
			files := maps.Clone(d.files)
			dirs := make([]string, 0, len(d.dirs))
			for name := range d.dirs {
				dirs = append(dirs, name)
			}
			testutil.WriteFiles(t, dir, files, dirs...)
			configFinder := policy.NewConfigFinder()
			p, err := configFinder.Find(testutil.Abs(dir, d.configFilePath), testutil.Abs(dir, d.wd))
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if exp := testutil.Abs(dir, d.exp); p != exp {
				t.Fatalf("wanted %v, got %v", exp, p)
			}
		})
	}
}
