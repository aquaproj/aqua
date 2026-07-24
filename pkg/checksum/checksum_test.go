package checksum_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
)

func TestCalculator_Calculate(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		content   string
		algorithm string
		checksum  string
		isErr     bool
	}{
		{
			name:  "algorithm is required",
			isErr: true,
		},
		{
			name:      "unsupported algorithm",
			isErr:     true,
			algorithm: pkgFoo,
		},
		{
			name:      "sha256",
			content:   helloWorld,
			algorithm: algoSHA256,
			checksum:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:      "sha512",
			content:   helloWorld,
			algorithm: algoSHA512,
			checksum:  "309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f",
		},
	}
	calculator := &checksum.Calculator{}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			filename := filepath.Join(t.TempDir(), "asset.tar.gz")
			if err := os.WriteFile(filename, []byte(d.content), osfile.FilePermission); err != nil {
				t.Fatal(err)
			}
			c, err := calculator.Calculate(filename, d.algorithm)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
			if c != d.checksum {
				t.Fatalf("wanted %s, got %s", d.checksum, c)
			}
		})
	}
}

func TestGetChecksumConfigFromFilename(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		filename string
		version  string
		want     *registry.Checksum
	}{
		{
			filename: "foo-1.0.0.tar.gz",
			version:  pkgVersion100,
			want:     nil,
		},
		{
			filename: "foo-1.0.0.tar.gz.sha256",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "foo-{{.Version}}.tar.gz.sha256",
			},
		},
		{
			filename: "sha1sums-1.0.0.txt",
			version:  pkgVersionV1,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "sha1",
				Asset:     "sha1sums-{{trimV .Version}}.txt",
			},
		},
		{
			filename: "sha1sums-1.0.0.asc",
			version:  pkgVersion100,
			want:     nil,
		},
		{
			filename: "SHA512SUMS",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA512,
				Asset:     "SHA512SUMS",
			},
		},
		{
			filename: "SHASUMS512.txt",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA512,
				Asset:     "SHASUMS512.txt",
			},
		},
		{
			filename: "SHASUMS256.txt",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "SHASUMS256.txt",
			},
		},
		{
			filename: "SHASUMS.txt",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "SHASUMS.txt",
			},
		},
		{
			filename: "SHA42SUMS.txt",
			version:  pkgVersion100,
			want:     nil,
		},
		{
			filename: "md5Sums.txt",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "md5",
				Asset:     "md5Sums.txt",
			},
		},
		{
			filename: "checksums",
			version:  pkgVersion100,
			want: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "checksums",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			got := checksum.GetChecksumConfigFromFilename(tt.filename, tt.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}
