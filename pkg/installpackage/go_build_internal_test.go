package installpackage

import (
	"context"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/osexec"
)

type spyExecutor struct {
	args []string
}

func (e *spyExecutor) ExecStderr(cmd *osexec.Cmd) (int, error) {
	e.args = cmd.Args
	return 0, nil
}

func (e *spyExecutor) ExecAndOutputWhenFailure(cmd *osexec.Cmd) (int, error) {
	e.args = cmd.Args
	return 0, nil
}

func TestGoBuildInstallerImpl_Install(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		buildTags []string
		exp       string
	}{
		{
			name: "no build tags",
			exp:  "go build -o /root/bin/foo ./cmd/foo",
		},
		{
			name:      "single build tag",
			buildTags: []string{"containers_image_openpgp"},
			exp:       "go build -tags containers_image_openpgp -o /root/bin/foo ./cmd/foo",
		},
		{
			name:      "multiple build tags are comma separated",
			buildTags: []string{"containers_image_openpgp", "exclude_graphdriver_btrfs"},
			exp:       "go build -tags containers_image_openpgp,exclude_graphdriver_btrfs -o /root/bin/foo ./cmd/foo",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			exec := &spyExecutor{}
			is := NewGoBuildInstallerImpl(exec)
			if err := is.Install(context.Background(), "/root/bin/foo", "/root/src", "./cmd/foo", d.buildTags); err != nil {
				t.Fatal(err)
			}
			// cmd.Args[0] is the resolved path of the go binary, so compare the rest.
			got := "go " + strings.Join(exec.args[1:], " ")
			if got != d.exp {
				t.Fatalf("got %q, wanted %q", got, d.exp)
			}
		})
	}
}
