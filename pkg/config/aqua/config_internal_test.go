package aqua

import (
	"testing"
)

func Test_parseNameWithVersion(t *testing.T) {
	t.Parallel()
	data := []struct {
		title      string
		name       string
		expName    string
		expVersion string
	}{
		{
			title:      "no version",
			name:       pkgFoo,
			expName:    pkgFoo,
			expVersion: "",
		},
		{
			title:      "with version",
			name:       "foo@v1.0.0",
			expName:    pkgFoo,
			expVersion: "v1.0.0",
		},
		{
			title:      "invalid name @v1.0.0",
			name:       "@v1.0.0",
			expName:    "",
			expVersion: "v1.0.0",
		},
		{
			title:      "invalid name foo@",
			name:       "foo@",
			expName:    pkgFoo,
			expVersion: "",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			name, version := parseNameWithVersion(d.name)
			if name != d.expName {
				t.Fatalf("name is got %s, wanted %s", name, d.expName)
			}
			if version != d.expVersion {
				t.Fatalf("version is got %s, wanted %s", version, d.expVersion)
			}
		})
	}
}
