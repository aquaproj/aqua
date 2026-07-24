package initpolicy_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/controller/initpolicy"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
)

func TestController_Init(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		files map[string]string
		param *config.Param
		isErr bool
	}{
		{
			name: "file already exists",
			files: map[string]string{
				"aqua-policy.yaml": `registries:
- type: standard
  ref: semver(">= 3.0.0")
packages:
- registry: standard
`,
			},
		},
		{
			name: "normal",
			param: &config.Param{
				CWD: "/home/foo/workspace",
			},
			files: map[string]string{},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files)
			ctrl := initpolicy.New()
			// The policy file is created relative to the working directory when
			// no path is given, so the test must give one.
			if err := ctrl.Init(logger, filepath.Join(dir, "aqua-policy.yaml")); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

// An empty path defaults to aqua-policy.yaml in the working directory. The test
// changes into a temporary directory to exercise that default without writing
// into the package directory, which is why it can't run in parallel.
func TestController_Init_defaultPath(t *testing.T) { //nolint:paralleltest // t.Chdir can't be used in a parallel test
	t.Chdir(t.TempDir())
	logger := slog.New(slog.DiscardHandler)
	ctrl := initpolicy.New()
	if err := ctrl.Init(logger, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("aqua-policy.yaml"); err != nil {
		t.Fatal(err)
	}
	// Running it again finds the existing file and does nothing.
	if err := ctrl.Init(logger, ""); err != nil {
		t.Fatal(err)
	}
}
