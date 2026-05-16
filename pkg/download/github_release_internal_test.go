package download

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/github"
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
					Name: new("foo"),
				},
			},
			assetName: "bar",
			isErr:     true,
		},
		{
			title: "found",
			assets: []*github.ReleaseAsset{
				{
					Name: new("foo"),
				},
				{
					Name: new("bar"),
					ID:   new(int64(10)),
				},
			},
			assetName: "bar",
			assetID:   10,
		},
	}
	for _, d := range data {
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

func Test_parseDigest(t *testing.T) { //nolint:funlen
	t.Parallel()
	const validHex = "3516a4d84f7b69ea5752ca2416895a2705910af3ed6815502af789000fc7e963"
	data := []struct {
		title  string
		digest string
		want   string // empty means nil expected
	}{
		{
			title:  "valid sha256",
			digest: "sha256:" + validHex,
			want:   "3516A4D84F7B69EA5752CA2416895A2705910AF3ED6815502AF789000FC7E963",
		},
		{
			title:  "valid sha256 uppercase hex is normalized",
			digest: "sha256:" + "ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789",
			want:   "ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789",
		},
		{
			title:  "missing colon",
			digest: "sha256" + validHex,
			want:   "",
		},
		{
			title:  "empty string",
			digest: "",
			want:   "",
		},
		{
			title:  "unsupported algorithm",
			digest: "sha512:" + validHex + validHex,
			want:   "",
		},
		{
			title:  "wrong length (too short)",
			digest: "sha256:abcd",
			want:   "",
		},
		{
			title:  "wrong length (too long)",
			digest: "sha256:" + validHex + "00",
			want:   "",
		},
		{
			title:  "non-hex character in payload",
			digest: "sha256:" + "g" + validHex[1:],
			want:   "",
		},
		{
			title:  "empty payload",
			digest: "sha256:",
			want:   "",
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			got := parseDigest(logger, d.digest)
			if d.want == "" {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %q, got nil", d.want)
			}
			if got.Digest != d.want {
				t.Fatalf("digest: wanted %q, got %q", d.want, got.Digest)
			}
			if got.Algorithm != "sha256" {
				t.Fatalf("algorithm: wanted %q, got %q", "sha256", got.Algorithm)
			}
		})
	}
}
