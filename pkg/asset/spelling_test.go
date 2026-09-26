package asset_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/asset"
)

func TestKnowsSpelling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		platform string
		spelling string
		want     bool
	}{
		{platform: "amd64", spelling: "x86_64", want: true},
		{platform: "arm64", spelling: "aarch64", want: true},
		{platform: "darwin", spelling: "apple-darwin", want: true},
		{platform: "darwin", spelling: "osx", want: true},
		{platform: "darwin", spelling: "macOS", want: true},
		{platform: "linux", spelling: "unknown-linux-musl", want: true},
		{platform: "windows", spelling: "pc-windows-msvc", want: true},
		{platform: "linux", spelling: "linux", want: true},
		// The one a definition has to say, because the release says nothing that tells
		// the parser this asset is a Linux build.
		{platform: "linux", spelling: "ubuntu", want: false},
		{platform: "darwin", spelling: "Darwin_universal", want: false},
		// A spelling the parser knows, but as another platform.
		{platform: "amd64", spelling: "aarch64", want: false},
		{platform: "linux", spelling: "osx", want: false},
		{platform: "", spelling: "osx", want: false},
		{platform: "linux", spelling: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.platform+"/"+tt.spelling, func(t *testing.T) {
			t.Parallel()
			if got := asset.KnowsSpelling(tt.platform, tt.spelling); got != tt.want {
				t.Errorf("KnowsSpelling(%q, %q) = %v, want %v", tt.platform, tt.spelling, got, tt.want)
			}
		})
	}
}
