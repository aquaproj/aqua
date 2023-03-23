package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestFileParser_ParseChecksumFile(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		content string
		pkg     *config.Package
		m       map[string]string
		s       string
		isErr   bool
	}{
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
		{
			name:    "sha256",
			content: `89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101  /home/runner/nova_3.2.0_darwin_arm64.tar.gz`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{
						FileFormat: "regexp",
						Pattern: &registry.ChecksumPattern{
							Checksum: `^(.{64})`,
						},
					},
				},
			},
			s: "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101",
		},
		{
			name:    "default",
			content: `89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101  nova_3.2.0_darwin_arm64.tar.gz`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{},
				},
			},
			m: map[string]string{
				"nova_3.2.0_darwin_arm64.tar.gz": "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101",
			},
		},
		{
			name:    "default absolute",
			content: `89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101  /home/runner/nova_3.2.0_darwin_arm64.tar.gz`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{},
				},
			},
			m: map[string]string{
				"nova_3.2.0_darwin_arm64.tar.gz": "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101",
			},
		},
		{
			name:    "default only checksum",
			content: `89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{},
				},
			},
			s: "89f744a88dad0e73866d06e79afccd5476152770c70101361566b234b0722101",
		},
		{
			name: "default multiple lines",
			content: `955ec1e8d329f75e6e3803ffae8a6f8e586576904ac418e9ed88f2cc31178c15  ./imgpkg-darwin-amd64
1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4  ./imgpkg-darwin-arm64
2c289cf6b5c88a4dd4bec17c9e57e49c2c7531c127ea130737945392cdc65362  ./imgpkg-linux-amd64
eb972061a7a71b03ee224b3e3d7aa0ec9a45ec20a6c8c5b11917b223c58a9570  ./imgpkg-linux-arm64
d3e8e4d8da6b6f5e0a77335864944fc3e74c109c3d4959c976c1caec1dc1807c  ./imgpkg-windows-amd64.exe
`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{},
				},
			},
			m: map[string]string{
				"imgpkg-darwin-amd64":      "955ec1e8d329f75e6e3803ffae8a6f8e586576904ac418e9ed88f2cc31178c15",
				"imgpkg-darwin-arm64":      "1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4",
				"imgpkg-linux-amd64":       "2c289cf6b5c88a4dd4bec17c9e57e49c2c7531c127ea130737945392cdc65362",
				"imgpkg-linux-arm64":       "eb972061a7a71b03ee224b3e3d7aa0ec9a45ec20a6c8c5b11917b223c58a9570",
				"imgpkg-windows-amd64.exe": "d3e8e4d8da6b6f5e0a77335864944fc3e74c109c3d4959c976c1caec1dc1807c",
			},
		},
	}
	parser := &checksum.FileParser{}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			m, s, err := parser.ParseChecksumFile(d.content, d.pkg.PackageInfo.Checksum)
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
			if s != d.s {
				t.Fatalf("wanted %s, got %s", d.s, s)
			}
		})
	}
}
