package runtime

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetRuntimesFromEnvs(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name  string
		envs  []string
		isErr bool
		rts   []*Runtime
	}{
		{
			name: "nil",
			rts:  allRuntimes(),
		},
		{
			name: "all",
			envs: []string{"darwin", "all"},
			rts:  allRuntimes(),
		},
		{
			name: "darwin amd64",
			envs: []string{"darwin", "amd64"},
			rts: []*Runtime{
				{
					GOOS:   "darwin",
					GOARCH: "amd64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
				{
					GOOS:   "windows",
					GOARCH: "amd64",
				},
				{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
			},
		},
		{
			name: "darwin linux/amd64",
			envs: []string{"darwin", "linux/amd64"},
			rts: []*Runtime{
				{
					GOOS:   "darwin",
					GOARCH: "amd64",
				},
				{
					GOOS:   "darwin",
					GOARCH: "arm64",
				},
				{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			rts, err := GetRuntimesFromEnvs(d.envs)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
			if diff := cmp.Diff(d.rts, rts); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
