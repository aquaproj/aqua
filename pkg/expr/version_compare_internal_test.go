package expr

import "testing"

func Test_compare(t *testing.T) {
	t.Parallel()
	data := []struct {
		name       string
		constraint string
		version    string
		exp        bool
	}{
		{
			name:       "invalid semver",
			constraint: ">= 4.0.0",
			version:    "35661968adb8fa29ab1d4a8713c0547d9a6007bb",
			exp:        false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := compare(d.constraint, d.version)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
