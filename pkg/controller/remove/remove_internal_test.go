package remove

import "testing"

func Test_parsePkgName(t *testing.T) {
	t.Parallel()
	data := []struct {
		name         string
		input        string
		registryName string
		pkgName      string
	}{
		{
			name:         "normal",
			input:        "cli/cli",
			registryName: "standard",
			pkgName:      "cli/cli",
		},
		{
			name:         "custom registry",
			input:        "foo,cli/cli",
			registryName: "foo",
			pkgName:      "cli/cli",
		},
	}

	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			r, p := parsePkgName(d.input)
			if r != d.registryName {
				t.Fatalf("registry name: wanted %s, got %s", d.registryName, r)
			}
			if p != d.pkgName {
				t.Fatalf("package name: wanted %s, got %s", d.pkgName, p)
			}
		})
	}
}
