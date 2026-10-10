package download

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/template"
)

func TestForgeReleaseURL(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title string
		typ   string
		host  string
		owner string
		repo  string
		tag   string
		asset string
		exp   string
	}{
		{
			title: "a Forgejo instance",
			typ:   config.PkgInfoTypeForgejoRelease,
			host:  "codeberg.org",
			owner: "mergiraf",
			repo:  "mergiraf",
			tag:   "v0.20.0",
			asset: "mergiraf_x86_64-unknown-linux-gnu.tar.gz",
			exp:   "https://codeberg.org/mergiraf/mergiraf/releases/download/v0.20.0/mergiraf_x86_64-unknown-linux-gnu.tar.gz",
		},
		{
			title: "a tag holding a slash",
			typ:   config.PkgInfoTypeGiteaRelease,
			host:  "codeberg.org",
			owner: "an-owner",
			repo:  "a-repo",
			tag:   "kustomize/v5.8.1",
			asset: "a-repo.tar.gz",
			exp:   "https://codeberg.org/an-owner/a-repo/releases/download/kustomize%2Fv5.8.1/a-repo.tar.gz",
		},
		{
			// GitLab serves it at a path of its own: the permanent release link.
			title: "a GitLab instance",
			typ:   config.PkgInfoTypeGitLabRelease,
			host:  "gitlab.com",
			owner: "gitlab-org",
			repo:  "cli",
			tag:   "v1.122.0",
			asset: "glab_1.122.0_linux_amd64.tar.gz",
			exp:   "https://gitlab.com/gitlab-org/cli/-/releases/v1.122.0/downloads/glab_1.122.0_linux_amd64.tar.gz",
		},
		{
			// A project in subgroups is named by its namespace, and those separators
			// say where the project is rather than being upstream's text.
			title: "a GitLab project in subgroups",
			typ:   config.PkgInfoTypeGitLabRelease,
			host:  "gitlab.com",
			owner: "gitlab-org/security",
			repo:  "cli",
			tag:   "v1.122.0",
			asset: "glab.tar.gz",
			exp:   "https://gitlab.com/gitlab-org/security/cli/-/releases/v1.122.0/downloads/glab.tar.gz",
		},
		{
			// An asset is named by the file path its release link was created with,
			// which can be nested: those separators say where the file is.
			title: "a GitLab asset that is a path",
			typ:   config.PkgInfoTypeGitLabRelease,
			host:  "gitlab.com",
			owner: "an-owner",
			repo:  "a-repo",
			tag:   "v1.0.0",
			asset: "binaries/linux-amd64",
			exp:   "https://gitlab.com/an-owner/a-repo/-/releases/v1.0.0/downloads/binaries/linux-amd64",
		},
		{
			title: "a GitLab tag holding a slash",
			typ:   config.PkgInfoTypeGitLabRelease,
			host:  "gitlab.com",
			owner: "an-owner",
			repo:  "a-repo",
			tag:   "cli/v1.0.0",
			asset: "a-repo.tar.gz",
			exp:   "https://gitlab.com/an-owner/a-repo/-/releases/cli%2Fv1.0.0/downloads/a-repo.tar.gz",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			u, err := forgeReleaseURL(d.typ, d.host, d.owner, d.repo, d.tag, d.asset)
			if err != nil {
				t.Fatal(err)
			}
			if u != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, u)
			}
		})
	}
}

func TestForgeReleaseURL_noInstance(t *testing.T) {
	t.Parallel()
	// Nothing says where to download from, which is a URL that must not be built: it
	// would be https:///an-owner/a-repo/..., a request to nowhere reported as a 404.
	if _, err := forgeReleaseURL(config.PkgInfoTypeForgejoRelease, "", "an-owner", "a-repo", "v1.0.0", "a-repo.tar.gz"); err == nil {
		t.Fatal("an error must be returned")
	}
}

// A type that is on no instance has no such URL, which is a caller asking the wrong
// question rather than a definition being wrong.
func TestForgeReleaseURL_unknownType(t *testing.T) {
	t.Parallel()
	if _, err := forgeReleaseURL(config.PkgInfoTypeHTTP, "an.example.com", "an-owner", "a-repo", "v1.0.0", "a-repo.tar.gz"); err == nil {
		t.Fatal("an error must be returned")
	}
}

// A file downloaded beside an asset -- a signature, say -- is on the instance that
// served the asset, which the package said and the file doesn't repeat.
func TestConvertDownloadedFileToFile_forgeRelease(t *testing.T) {
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

// Each type is downloaded from the instance the package is on, and a package of any
// is read as the release it says it is.
func TestConvertPackageToFile_forge(t *testing.T) {
	t.Parallel()
	for _, typ := range []string{config.PkgInfoTypeForgejoRelease, config.PkgInfoTypeGiteaRelease, config.PkgInfoTypeGitLabRelease} {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			file, err := ConvertPackageToFile(&config.Package{
				PackageInfo: &registry.PackageInfo{
					Type:      typ,
					Host:      "an.example.com",
					RepoOwner: "an-owner",
					RepoName:  "a-repo",
					Asset:     "a-repo.tar.gz",
				},
				Package: &aqua.Package{Version: "v1.0.0"},
			}, "a-repo.tar.gz", &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"})
			if err != nil {
				t.Fatal(err)
			}
			if file.Type != typ || file.Host != "an.example.com" || file.Asset != "a-repo.tar.gz" {
				t.Fatalf("wanted the release the package says it is, got %+v", file)
			}
		})
	}
}
