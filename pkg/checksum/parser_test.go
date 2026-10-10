package checksum_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
)

func TestParseChecksumFile(t *testing.T) { //nolint:funlen
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
			name:    algoSHA256,
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
				assetNova320DarwinArm64: checksumValue,
			},
		},
		{
			name:    algoSHA256,
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
			s: checksumValue,
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
				assetNova320DarwinArm64: checksumValue,
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
				// Where the file was built, and the file the release
				// publishes, are both what a line can be read as.
				"/home/runner/" + assetNova320DarwinArm64: checksumValue,
				assetNova320DarwinArm64:                   checksumValue,
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
			s: checksumValue,
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
				"./imgpkg-darwin-amd64":      "955ec1e8d329f75e6e3803ffae8a6f8e586576904ac418e9ed88f2cc31178c15",
				"./imgpkg-darwin-arm64":      "1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4",
				"./imgpkg-linux-amd64":       "2c289cf6b5c88a4dd4bec17c9e57e49c2c7531c127ea130737945392cdc65362",
				"./imgpkg-linux-arm64":       "eb972061a7a71b03ee224b3e3d7aa0ec9a45ec20a6c8c5b11917b223c58a9570",
				"./imgpkg-windows-amd64.exe": "d3e8e4d8da6b6f5e0a77335864944fc3e74c109c3d4959c976c1caec1dc1807c",
				"imgpkg-darwin-amd64":        "955ec1e8d329f75e6e3803ffae8a6f8e586576904ac418e9ed88f2cc31178c15",
				"imgpkg-darwin-arm64":        "1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4",
				"imgpkg-linux-amd64":         "2c289cf6b5c88a4dd4bec17c9e57e49c2c7531c127ea130737945392cdc65362",
				"imgpkg-linux-arm64":         "eb972061a7a71b03ee224b3e3d7aa0ec9a45ec20a6c8c5b11917b223c58a9570",
				"imgpkg-windows-amd64.exe":   "d3e8e4d8da6b6f5e0a77335864944fc3e74c109c3d4959c976c1caec1dc1807c",
			},
		},
		{
			// Two paths ending in the same file name: nothing is published as
			// that name, and each path is what its own line answers for.
			name: "default the same file name in two places",
			content: `955ec1e8d329f75e6e3803ffae8a6f8e586576904ac418e9ed88f2cc31178c15  binaries/linux-amd64/tool
1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4  binaries/darwin-arm64/tool
`,
			pkg: &config.Package{
				PackageInfo: &registry.PackageInfo{
					Checksum: &registry.Checksum{},
				},
			},
			m: map[string]string{
				"binaries/linux-amd64/tool":  "955ec1e8d329f75e6e3803ffae8a6f8e586576904ac418e9ed88f2cc31178c15",
				"binaries/darwin-arm64/tool": "1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4",
			},
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			m, s, err := checksum.ParseChecksumFile(d.content, d.pkg.PackageInfo.Checksum)
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

// An asset named by a path is what a release link on a GitLab instance can publish, and
// the checksum file names it either way.
func TestFindChecksum(t *testing.T) {
	t.Parallel()
	m := map[string]string{
		"binaries/linux-amd64/tool":      checksumValue,
		"nova_3.2.0_darwin_arm64.tar.gz": "1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4",
	}
	data := []struct {
		name  string
		asset string
		want  string
		found bool
	}{
		{
			name:  "the path the file wrote",
			asset: "binaries/linux-amd64/tool",
			want:  checksumValue,
			found: true,
		},
		{
			name:  "a path the file names by the file alone",
			asset: "dist/nova_3.2.0_darwin_arm64.tar.gz",
			want:  "1182217e827a44e22df22be4b95e5392f52530eaed52da5195b5f026d06f41f4",
			found: true,
		},
		{
			name:  "the file alone, of a path the file wrote",
			asset: "tool",
		},
		{
			name:  "an asset the file says nothing about",
			asset: "nova_3.2.0_linux_amd64.tar.gz",
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			chksum, found := checksum.FindChecksum(m, d.asset)
			if found != d.found {
				t.Fatalf("wanted found to be %v, got %v", d.found, found)
			}
			if chksum != d.want {
				t.Fatalf("wanted %s, got %s", d.want, chksum)
			}
		})
	}
}
