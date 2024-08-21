package outputshell

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

func Test_getNewPS(t *testing.T) {
	t.Parallel()
	data := []struct {
		name     string
		param    *config.Param
		shell    *Shell
		oldPaths map[string]struct{}
		ps       []string
		updated  bool
	}{
		{
			name: "normal",
			param: &config.Param{
				EnvPath:           "/root/.local/share/aquaproj-aqua/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				PathListSeparator: ":",
			},
			shell: &Shell{
				Env: &Env{
					Path: &Path{
						Values: []string{
							"/root/.local/share/aquaproj-aqua/pkgs/http/nodejs.org/dist/v20.16.0/node-v20.16.0-linux-arm64.tar.gz/node-v20.16.0-linux-arm64/bin",
						},
					},
				},
			},
			oldPaths: map[string]struct{}{},
			ps: []string{
				"/root/.local/share/aquaproj-aqua/pkgs/http/nodejs.org/dist/v20.16.0/node-v20.16.0-linux-arm64.tar.gz/node-v20.16.0-linux-arm64/bin",
				"/root/.local/share/aquaproj-aqua/bin",
				"/usr/local/sbin",
				"/usr/local/bin",
				"/usr/sbin",
				"/usr/bin",
				"/sbin",
				"/bin",
			},
			updated: true,
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
