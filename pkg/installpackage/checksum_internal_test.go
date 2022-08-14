package installpackage

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/runtime"
)

func strP(s string) *string {
	return &s
}

func TestInstaller_extractChecksum(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name         string
		pkg          *config.Package
		runtime      *runtime.Runtime
		assetName    string
		checksumFile []byte
		isErr        bool
		c            string
	}{
		{
			name: "raw",
			runtime: &runtime.Runtime{
				GOOS:   "windows",
				GOARCH: "amd64",
			},
			assetName:    "spotify-tui-windows.sha256",
			checksumFile: []byte("423b7c1a842bb4cd847b492cad2d1c724e00f92ed913226e9ac1ab0925b0b639"),
			c:            "423b7c1a842bb4cd847b492cad2d1c724e00f92ed913226e9ac1ab0925b0b639",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "Rigellute/spotify-tui",
					Version: "v0.25.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "Rigellute",
					RepoName:  "spotify-tui",
					Asset:     strP("spotify-tui-{{.OS}}.tar.gz"),
					Checksum: &registry.Checksum{
						Type:       "github_release",
						FileFormat: "raw",
						Asset:      "spotify-tui-{{.OS}}.sha256",
						Algorithm:  "sha256",
					},
				},
			},
		},
		{
			name: "colima",
			runtime: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			assetName:    "colima-Darwin-arm64",
			checksumFile: []byte("8d5b5f66ec04a4e21d147e2b7501d44b084eacf3dc03af081e681fb99697c3ac  colima-Darwin-arm64\n"),
			c:            "8d5b5f66ec04a4e21d147e2b7501d44b084eacf3dc03af081e681fb99697c3ac",
			pkg: &config.Package{
				Package: &aqua.Package{
					Name:    "abiosoft/colima",
					Version: "v0.4.4",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "abiosoft",
					RepoName:  "colima",
					Asset:     strP("colima-{{.OS}}-{{.Arch}}"),
					Replacements: map[string]string{
						"darwin": "Darwin",
						"linux":  "Linux",
						"amd64":  "x86_64",
					},
					Checksum: &registry.Checksum{
						Type:       "github_release",
						FileFormat: "regexp",
						Asset:      "colima-{{.OS}}-{{.Arch}}.sha256sum",
						Algorithm:  "sha256",
						Pattern: &registry.ChecksumPattern{
							Checksum: `^(.{64})`,
							File:     `^.{64}\s+(\S+)$`,
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			inst := &Installer{
				runtime: &runtime.Runtime{
					GOOS:   "linux",
					GOARCH: "amd64",
				},
			}
			if d.runtime != nil {
				inst.runtime = d.runtime
			}
			c, err := inst.extractChecksum(d.pkg, d.assetName, d.checksumFile)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must occur")
			}
			if c != d.c {
				t.Fatalf("checksum wanted %s, got %s", d.c, c)
			}
		})
	}
}
