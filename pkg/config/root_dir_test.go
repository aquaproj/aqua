package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func TestGetRootDir(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		env  map[string]string
		exp  string
	}{
		{
			name: "AQUA_ROOT_DIR",
			env: map[string]string{
				"AQUA_ROOT_DIR": "/home/foo/.aqua",
			},
			exp: "/home/foo/.aqua",
		},
		{
			name: "XDG_DATA_HOME",
			env: map[string]string{
				"XDG_DATA_HOME": "/home/foo/.xdg",
			},
			exp: "/home/foo/.xdg/aquaproj-aqua",
		},
		{
			name: "HOME",
			env: map[string]string{
				"HOME": "/home/foo",
			},
			exp: "/home/foo/.local/share/aquaproj-aqua",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			rootDir := config.GetRootDir(osenv.NewMock(d.env))
			if rootDir != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, rootDir)
			}
		})
	}
}
