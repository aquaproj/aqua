package osfile_test

import (
	"os"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

func TestIsOwnerExecutable(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		mode os.FileMode
		exp  bool
	}{
		{
			name: "true",
			mode: 0o100,
			exp:  true,
		},
		{
			name: "false",
			mode: 0o200,
			exp:  false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			f := osfile.IsOwnerExecutable(d.mode)
			if f != d.exp {
				t.Fatalf("watnted %v, got %v", d.exp, f)
			}
		})
	}
}

func TestAllowOwnerExec(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		mode os.FileMode
		exp  os.FileMode
	}{
		{
			name: "true",
			mode: 0o100,
			exp:  0o100,
		},
		{
			name: "false",
			mode: 0o200,
			exp:  0o300,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			mode := osfile.AllowOwnerExec(d.mode)
			if mode != d.exp {
				t.Fatalf("watnted %v, got %v", d.exp, mode)
			}
		})
	}
}
