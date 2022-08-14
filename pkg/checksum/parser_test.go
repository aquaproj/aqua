package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestFileParser_ParseChecksumFile(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		content string
		pkg     *config.Package
		m       map[string]string
		isErr   bool
	}{
		{
			name: "unknown format",
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{},
				},
			},
			isErr: true,
		},
		{
			name:    "sha256",
			content: `89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101  nova_3.2.0_darwin_arm64.tar.gz`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{
						FileFormat: "regexp",
						Pattern: &registry.ChecksumPattern{
							Checksum: `^(.{64})`,
							File:     `^.{64}\s+(\S+)$`,
						},
					},
				},
			},
			m: map[string]string{
				"nova_3.2.0_darwin_arm64.tar.gz": "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101",
			},
		},
	}
	parser := &checksum.FileParser{}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			m, err := parser.ParseChecksumFile(d.content, d.pkg)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
			if diff := cmp.Diff(m, d.m); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
