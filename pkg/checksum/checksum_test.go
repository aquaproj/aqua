package checksum_test

import (
	"reflect"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
)

func TestCalculator_Calculate(t *testing.T) {
	t.Parallel()
	data := []struct {
		name      string
		filename  string
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
			algorithm: "foo",
		},
	}
	calculator := &checksum.Calculator{}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			if err := afero.WriteFile(fs, d.filename, []byte(d.content), osfile.FilePermission); err != nil {
				t.Fatal(err)
			}
			c, err := calculator.Calculate(fs, d.filename, d.algorithm)
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
			version:  "1.0.0",
			want:     nil,
		},
		{
			filename: "foo-1.0.0.tar.gz.sha256",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha256",
				Asset:     "foo-{{.Version}}.tar.gz.sha256",
			},
		},
		{
			filename: "sha1sums-1.0.0.txt",
			version:  "v1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha1",
				Asset:     "sha1sums-{{trimV .Version}}.txt",
			},
		},
		{
			filename: "sha1sums-1.0.0.asc",
			version:  "1.0.0",
			want:     nil,
		},
		{
			filename: "SHA512SUMS",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha512",
				Asset:     "SHA512SUMS",
			},
		},
		{
			filename: "SHASUMS512.txt",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha512",
				Asset:     "SHASUMS512.txt",
			},
		},
		{
			filename: "SHASUMS256.txt",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha256",
				Asset:     "SHASUMS256.txt",
			},
		},
		{
			filename: "SHASUMS.txt",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha256",
				Asset:     "SHASUMS.txt",
			},
		},
		{
			filename: "SHA42SUMS.txt",
			version:  "1.0.0",
			want:     nil,
		},
		{
			filename: "md5Sums.txt",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "md5",
				Asset:     "md5Sums.txt",
			},
		},
		{
			filename: "checksums",
			version:  "1.0.0",
			want: &registry.Checksum{
				Type:      "github_release",
				Algorithm: "sha256",
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
