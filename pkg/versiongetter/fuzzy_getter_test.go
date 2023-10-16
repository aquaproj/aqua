package versiongetter_test

import (
	"context"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter"
	"github.com/sirupsen/logrus"
)

func TestFuzzyGetter_Get(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name           string
		pkg            *registry.PackageInfo
		currentVersion string
		useFinder      bool
		version        string
		idxs           []int
		versions       map[string][]*fuzzyfinder.Item
	}{
		{
			name: "normal",
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			currentVersion: "v2.0.0",
			version:        "v4.6.0",
			versions: map[string][]*fuzzyfinder.Item{
				"suzuki-shunsuke/tfcmt": {
					{
						Item: "v4.6.0",
					},
					{
						Item: "v3.0.0",
					},
					{
						Item: "v2.0.0",
					},
				},
			},
		},
		{
			name: "finder",
			pkg: &registry.PackageInfo{
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "tfcmt",
			},
			useFinder:      true,
			idxs:           []int{1},
			currentVersion: "v2.0.0",
			version:        "v3.0.0",
			versions: map[string][]*fuzzyfinder.Item{
				"suzuki-shunsuke/tfcmt": {
					{
						Item: "v4.6.0",
					},
					{
						Item: "v3.0.0",
					},
					{
						Item: "v2.0.0",
					},
				},
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			finder := fuzzyfinder.NewMock(d.idxs, nil)
			vg := versiongetter.NewMockVersionGetter(d.versions)
			fg := versiongetter.NewFuzzy(finder, vg)
			version := fg.Get(context.Background(), logE, d.pkg, d.currentVersion, d.useFinder)
			if version != d.version {
				t.Fatalf("wanted %s, got %s", d.version, version)
			}
		})
	}
}
