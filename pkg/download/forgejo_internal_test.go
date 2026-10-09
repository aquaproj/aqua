package download

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

func TestForgejoReleaseURL(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		host  string
		owner string
		repo  string
		tag   string
		asset string
		exp   string
	}{
		{
			title: "normal",
			host:  "codeberg.org",
			owner: "mergiraf",
			repo:  "mergiraf",
			tag:   "v0.20.0",
			asset: "mergiraf_x86_64-unknown-linux-gnu.tar.gz",
			exp:   "https://codeberg.org/mergiraf/mergiraf/releases/download/v0.20.0/mergiraf_x86_64-unknown-linux-gnu.tar.gz",
		},
		{
			title: "a tag holding a slash",
			host:  "codeberg.org",
			owner: "an-owner",
			repo:  "a-repo",
			tag:   "kustomize/v5.8.1",
			asset: "a-repo.tar.gz",
			exp:   "https://codeberg.org/an-owner/a-repo/releases/download/kustomize%2Fv5.8.1/a-repo.tar.gz",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if u := forgejoReleaseURL(d.host, d.owner, d.repo, d.tag, d.asset); u != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, u)
			}
		})
	}
}

// A file downloaded beside an asset -- a signature, say -- is on the instance that
// served the asset, which the package said and the file doesn't repeat.
func TestConvertDownloadedFileToFile_forgejoRelease(t *testing.T) {
	t.Parallel()
	asset := "mergiraf_x86_64-unknown-linux-gnu.tar.gz.minisig"
	file, err := ConvertDownloadedFileToFile(
		&registry.DownloadedFile{
			Type:  config.PkgInfoTypeForgejoRelease,
			Asset: &asset,
		},
		&File{
			Host:      "codeberg.org",
			RepoOwner: "mergiraf",
			RepoName:  "mergiraf",
			Version:   "v0.20.0",
		},
		&runtime.Runtime{GOOS: "linux", GOARCH: "amd64"},
		&template.Artifact{Version: "v0.20.0"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if file.Host != "codeberg.org" || file.RepoOwner != "mergiraf" || file.RepoName != "mergiraf" {
		t.Fatalf("wanted the package's instance and repository, got %+v", file)
	}
	if file.Asset != asset {
		t.Fatalf("wanted %s, got %s", asset, file.Asset)
	}
	if file.Version != "v0.20.0" {
		t.Fatalf("wanted the package's version, got %s", file.Version)
	}
}
