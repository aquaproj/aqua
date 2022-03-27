package unarchive_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/unarchive"
)

func TestIsUnarchived(t *testing.T) {
	t.Parallel()
	data := []struct {
		title       string
		archiveType string
		assetName   string
		exp         bool
	}{
		{
			title:     "tar.gz",
			assetName: "foo.tar.gz",
			exp:       false,
		},
		{
			title:     "archive is empty and assetName has no extension",
			assetName: "foo",
			exp:       true,
		},
		{
			title:       "archiveType is raw",
			assetName:   "foo-v3.0.0",
			archiveType: "raw",
			exp:         true,
		},
		{
			title:       "archiveTyps is set and isn't raw",
			assetName:   "foo",
			archiveType: "tar.gz",
			exp:         false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			f := unarchive.IsUnarchived(d.archiveType, d.assetName)
			if f != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, f)
			}
		})
	}
}
