package registry_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestPackageInfos_ToMap(t *testing.T) {
	t.Parallel()
	data := []struct {
		title    string
		pkgInfos *registry.PackageInfos
		exp      map[string]*registry.PackageInfo
	}{
		{
			title: "normal",
			pkgInfos: &registry.PackageInfos{
				&registry.PackageInfo{
					Type: "github_release",
					Name: "foo",
				},
			},
			exp: map[string]*registry.PackageInfo{
				"foo": {
					Type: "github_release",
					Name: "foo",
				},
			},
		},
	}

	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			m := d.pkgInfos.ToMap(logger)
			if diff := cmp.Diff(d.exp, m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
