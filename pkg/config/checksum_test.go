package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

func strP(s string) *string {
	return &s
}

func TestPackage_GetChecksumID(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		pkg        *config.Package
		checksumID string
		isErr      bool
		rt         *runtime.Runtime
	}{
		{
			name: "github_archive",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v3.0.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_archive",
					RepoOwner: "tfutils",
					RepoName:  "tfenv",
				},
			},
			rt:         &runtime.Runtime{},
			checksumID: "github_archive/github.com/tfutils/tfenv/v3.0.0",
		},
		{
			name: "github_content",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v1.1.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-installer",
					Path:      strP("aqua-installer"),
				},
			},
			rt:         &runtime.Runtime{},
			checksumID: "github_content/github.com/aquaproj/aqua-installer/v1.1.0/aqua-installer",
		},
		{
			name: "github_release",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v2.17.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     strP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Format:    "tar.gz",
					Replacements: map[string]string{
						"darwin": "macOS",
					},
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
			checksumID: "github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_macOS_amd64.tar.gz",
		},
		{
			name: "http",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v1.3.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "http",
					RepoOwner: "hashicorp",
					RepoName:  "terrafrom",
					URL:       strP("https://releases.hashicorp.com/terraform/{{trimV .Version}}/terraform_{{trimV .Version}}_{{.OS}}_{{.Arch}}.zip"),
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "amd64",
			},
			checksumID: "http/releases.hashicorp.com/terraform/1.3.0/terraform_1.3.0_darwin_amd64.zip",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksumID, err := d.pkg.GetChecksumID(d.rt)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if checksumID != d.checksumID {
				t.Fatalf("wanted %s, got %s", d.checksumID, checksumID)
			}
		})
	}
}

func TestPackage_GetChecksumIDFromAsset(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name       string
		pkg        *config.Package
		checksumID string
		isErr      bool
		asset      string
	}{
		{
			name: "github_archive",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v3.0.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_archive",
					RepoOwner: "tfutils",
					RepoName:  "tfenv",
				},
			},
			checksumID: "github_archive/github.com/tfutils/tfenv/v3.0.0",
		},
		{
			name: "github_content",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v1.1.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_content",
					RepoOwner: "aquaproj",
					RepoName:  "aqua-installer",
				},
			},
			checksumID: "github_content/github.com/aquaproj/aqua-installer/v1.1.0/aqua-installer",
			asset:      "aqua-installer",
		},
		{
			name: "github_release",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v2.17.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
				},
			},
			checksumID: "github_release/github.com/cli/cli/v2.17.0/gh_2.17.0_macOS_amd64.tar.gz",
			asset:      "gh_2.17.0_macOS_amd64.tar.gz",
		},
		{
			name: "http",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v1.3.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "http",
					RepoOwner: "hashicorp",
					RepoName:  "terrafrom",
					URL:       strP("https://releases.hashicorp.com/terraform/{{trimV .Version}}/terraform_{{trimV .Version}}_{{.OS}}_{{.Arch}}.zip"),
				},
			},
			checksumID: "http/releases.hashicorp.com/terraform/1.3.0/terraform_1.3.0_darwin_amd64.zip",
			asset:      "terraform_1.3.0_darwin_amd64.zip",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksumID, err := d.pkg.GetChecksumIDFromAsset(d.asset)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if checksumID != d.checksumID {
				t.Fatalf("wanted %s, got %s", d.checksumID, checksumID)
			}
		})
	}
}

func TestPackage_RenderChecksumFileName(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name             string
		pkg              *config.Package
		checksumFileName string
		isErr            bool
		rt               *runtime.Runtime
	}{
		{
			name: "github_release",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v2.17.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "github_release",
					RepoOwner: "cli",
					RepoName:  "cli",
					Asset:     strP("gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}"),
					Checksum: &registry.Checksum{
						Type:  "github_release",
						Asset: "gh_{{trimV .Version}}_checksums.txt",
					},
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			checksumFileName: "gh_2.17.0_checksums.txt",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			checksumFileName, err := d.pkg.RenderChecksumFileName(d.rt)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if checksumFileName != d.checksumFileName {
				t.Fatalf("wanted %s, got %s", d.checksumFileName, checksumFileName)
			}
		})
	}
}

func TestPackage_RenderChecksumURL(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name  string
		pkg   *config.Package
		url   string
		isErr bool
		rt    *runtime.Runtime
	}{
		{
			name: "normal",
			pkg: &config.Package{
				Package: &aqua.Package{
					Version: "v3.10.0",
				},
				PackageInfo: &registry.PackageInfo{
					Type:      "http",
					RepoOwner: "helm",
					RepoName:  "helm",
					URL:       strP("https://get.helm.sh/helm-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz"),
					Checksum: &registry.Checksum{
						Type: "http",
						URL:  "https://get.helm.sh/helm-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz.sha256sum",
					},
				},
			},
			rt: &runtime.Runtime{
				GOOS:   "darwin",
				GOARCH: "arm64",
			},
			url: "https://get.helm.sh/helm-v3.10.0-darwin-arm64.tar.gz.sha256sum",
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			u, err := d.pkg.RenderChecksumURL(d.rt)
			if err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
			if u != d.url {
				t.Fatalf("wanted %s, got %s", d.url, u)
			}
		})
	}
}
