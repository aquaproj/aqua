package finder_test

import (
	"reflect"
	"testing"

	finder "github.com/aquaproj/aqua/pkg/config-finder"
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
