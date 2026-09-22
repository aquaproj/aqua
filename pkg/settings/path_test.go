package settings_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/settings"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func TestPath(t *testing.T) {
	t.Parallel()
	data := []struct {
		name string
		env  map[string]string
		exp  string
	}{
		{
			name: "XDG_CONFIG_HOME",
			env: map[string]string{
				"XDG_CONFIG_HOME": "/home/foo/.xdg",
			},
			exp: "/home/foo/.xdg/aquaproj-aqua/config.yaml",
		},
		{
			name: "HOME",
			env: map[string]string{
				"HOME": "/home/foo",
			},
			exp: "/home/foo/.config/aquaproj-aqua/config.yaml",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			p := settings.Path(osenv.NewMock(d.env))
			if p != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, p)
			}
		})
	}
}
