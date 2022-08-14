package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/spf13/afero"
)

func TestCalculate(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		filename  string
		content   string
		algorithm string
		checksum  string
		isErr     bool
	}{
		{
			name:  "algorithm is required",
			isErr: true,
		},
		{
			name:      "unsupported algorithm",
			isErr:     true,
			algorithm: "foo",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			if err := afero.WriteFile(fs, d.filename, []byte(d.content), 0o644); err != nil {
				t.Fatal(err)
			}
			c, err := checksum.Calculate(fs, d.filename, d.algorithm)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
			if c != d.checksum {
				t.Fatalf("wanted %s, got %s", d.checksum, c)
			}
		})
	}
}
