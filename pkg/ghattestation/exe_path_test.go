package ghattestation_test

import (
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

// The path has to be the one the installer writes to, since one caller installs the
// tool and another runs it.
func TestExePath(t *testing.T) {
	t.Parallel()
	v := strings.TrimPrefix(ghattestation.Version, "v")
	for _, d := range []struct {
		name string
		rt   *runtime.Runtime
		want string
	}{
		{
			name: "linux",
			rt:   &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
			want: "/root/pkgs/github_release/github.com/cli/cli/" + ghattestation.Version + "/gh_" + v + "_linux_amd64.tar.gz/gh_" + v + "_linux_amd64/bin/gh",
		},
		{
			name: "windows",
			rt:   &runtime.Runtime{GOOS: "windows", GOARCH: "amd64"},
			want: "/root/pkgs/github_release/github.com/cli/cli/" + ghattestation.Version + "/gh_" + v + "_windows_amd64.zip/bin/gh.exe",
		},
	} {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			got, err := ghattestation.ExePath(&ghattestation.ParamExePath{RootDir: "/root", Runtime: d.rt})
			if err != nil {
				t.Fatal(err)
			}
			if got != d.want {
				t.Errorf("got %q, want %q", got, d.want)
			}
		})
	}
}
