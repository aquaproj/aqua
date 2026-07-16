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
			algorithm: pkgFoo,
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
		name         string
		assets       []string
		version      string
		wantName     string
		wantChecksum *registry.Checksum
	}{
		{
			name:         "foo-1.0.0.tar.gz",
			assets:       []string{"foo-1.0.0.tar.gz"},
			version:      pkgVersion100,
			wantName:     "",
			wantChecksum: nil,
		},
		{
			name:     "foo-1.0.0.tar.gz.sha256",
			assets:   []string{"foo-1.0.0.tar.gz.sha256"},
			version:  pkgVersion100,
			wantName: "foo-1.0.0.tar.gz.sha256",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "foo-{{.Version}}.tar.gz.sha256",
			},
		},
		{
			name:     "sha1sums-1.0.0.txt",
			assets:   []string{"sha1sums-1.0.0.txt"},
			version:  pkgVersionV1,
			wantName: "sha1sums-1.0.0.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "sha1",
				Asset:     "sha1sums-{{trimV .Version}}.txt",
			},
		},
		{
			name:         "sha1sums-1.0.0.asc",
			assets:       []string{"sha1sums-1.0.0.asc"},
			version:      pkgVersion100,
			wantName:     "",
			wantChecksum: nil,
		},
		{
			name:     "SHA512SUMS",
			assets:   []string{"SHA512SUMS"},
			version:  pkgVersion100,
			wantName: "SHA512SUMS",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA512,
				Asset:     "SHA512SUMS",
			},
		},
		{
			name:     "SHASUMS512.txt",
			assets:   []string{"SHASUMS512.txt"},
			version:  pkgVersion100,
			wantName: "SHASUMS512.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA512,
				Asset:     "SHASUMS512.txt",
			},
		},
		{
			name:     "SHASUMS256.txt",
			assets:   []string{"SHASUMS256.txt"},
			version:  pkgVersion100,
			wantName: "SHASUMS256.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "SHASUMS256.txt",
			},
		},
		{
			name:     "SHASUMS.txt",
			assets:   []string{"SHASUMS.txt"},
			version:  pkgVersion100,
			wantName: "SHASUMS.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "SHASUMS.txt",
			},
		},
		{
			name:         "SHA42SUMS.txt",
			assets:       []string{"SHA42SUMS.txt"},
			version:      pkgVersion100,
			wantName:     "",
			wantChecksum: nil,
		},
		{
			name:     "md5Sums.txt",
			assets:   []string{"md5Sums.txt"},
			version:  pkgVersion100,
			wantName: "md5Sums.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "md5",
				Asset:     "md5Sums.txt",
			},
		},
		{
			name:     "checksums",
			assets:   []string{"checksums"},
			version:  pkgVersion100,
			wantName: "checksums",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "checksums",
			},
		},
		// Pattern specificity: the generic sha256 words ("shasums", "checksum") are a
		// fallback matched last, so more specific algorithms must win when both appear.
		{
			name:     "checksums.md5",
			assets:   []string{"checksums.md5"},
			version:  pkgVersion100,
			wantName: "checksums.md5",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "md5",
				Asset:     "checksums.md5",
			},
		},
		{
			name:     "checksums.sha1",
			assets:   []string{"checksums.sha1"},
			version:  pkgVersion100,
			wantName: "checksums.sha1",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "sha1",
				Asset:     "checksums.sha1",
			},
		},
		{
			name:     "foo-1.0.0-checksums.md5",
			assets:   []string{"foo-1.0.0-checksums.md5"},
			version:  pkgVersion100,
			wantName: "foo-1.0.0-checksums.md5",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "md5",
				Asset:     "foo-{{.Version}}-checksums.md5",
			},
		},
		{
			name:     "SHASUMS.sha1",
			assets:   []string{"SHASUMS.sha1"},
			version:  pkgVersion100,
			wantName: "SHASUMS.sha1",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "sha1",
				Asset:     "SHASUMS.sha1",
			},
		},
		// When several candidate checksum files are provided, the one with the most
		// preferred algorithm is selected: sha512 > sha256 > sha1 > md5. The returned
		// filename is the original matched asset name.
		{
			name:     "sha512 is preferred over sha256",
			assets:   []string{"SHA256SUMS", "SHA512SUMS"},
			version:  pkgVersion100,
			wantName: "SHA512SUMS",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA512,
				Asset:     "SHA512SUMS",
			},
		},
		{
			name:     "sha256 is preferred over sha1",
			assets:   []string{"sha1sums.txt", "SHASUMS256.txt"},
			version:  pkgVersion100,
			wantName: "SHASUMS256.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "SHASUMS256.txt",
			},
		},
		{
			name:     "sha1 is preferred over md5",
			assets:   []string{"md5sums.txt", "sha1sums.txt"},
			version:  pkgVersion100,
			wantName: "sha1sums.txt",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: "sha1",
				Asset:     "sha1sums.txt",
			},
		},
		{
			name:     "preference is independent of input order",
			assets:   []string{"SHA256SUMS", "SHA512SUMS", "sha1sums.txt"},
			version:  pkgVersion100,
			wantName: "SHA512SUMS",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA512,
				Asset:     "SHA512SUMS",
			},
		},
		{
			name:     "signature files are ignored",
			assets:   []string{"SHA512SUMS.asc", "SHA256SUMS"},
			version:  pkgVersion100,
			wantName: "SHA256SUMS",
			wantChecksum: &registry.Checksum{
				Type:      pkgTypeGitHubRelease,
				Algorithm: algoSHA256,
				Asset:     "SHA256SUMS",
			},
		},
		{
			name:         "no checksum file returns nil",
			assets:       []string{"foo-1.0.0.tar.gz", "README.md"},
			version:      pkgVersion100,
			wantName:     "",
			wantChecksum: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotName, got := checksum.GetChecksumConfigFromFilename(tt.assets, tt.version)
			if gotName != tt.wantName {
				t.Errorf("wanted name %q, got %q", tt.wantName, gotName)
			}
			if !reflect.DeepEqual(got, tt.wantChecksum) {
				t.Errorf("wanted checksum %v, got %v", tt.wantChecksum, got)
			}
		})
	}
}
