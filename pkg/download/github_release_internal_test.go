package download

import (
	"testing"

	"github.com/google/go-github/v45/github"
)

func Test_getAssetIDFromAssets(t *testing.T) {
	t.Parallel()
	data := []struct {
		title     string
		assets    []*github.ReleaseAsset
		assetName string
		assetID   int64
		isErr     bool
	}{
		{
			title: "not found",
			assets: []*github.ReleaseAsset{
				{
					Name: github.String("foo"),
				},
			},
			assetName: "bar",
			isErr:     true,
		},
		{
			title: "found",
			assets: []*github.ReleaseAsset{
				{
					Name: github.String("foo"),
				},
				{
					Name: github.String("bar"),
					ID:   github.Int64(10),
				},
			},
			assetName: "bar",
			assetID:   10,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			assetID, err := getAssetIDFromAssets(d.assets, d.assetName)
			if d.isErr {
				if err == nil {
					t.Fatal("error should be returned")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if d.assetID != assetID {
				t.Fatalf("wanted %v, got %v", d.assetID, assetID)
			}
		})
	}
}
