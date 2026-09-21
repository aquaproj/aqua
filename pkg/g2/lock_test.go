package g2_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/g2"
	"github.com/aquaproj/aqua/v2/pkg/lockfile"
	"github.com/google/go-cmp/cmp"
)

func testRegistry() *g2.Registry {
	return &g2.Registry{
		Assets: []*g2.Asset{
			{
				OS:                "darwin",
				Arch:              "arm64",
				Type:              "github_release",
				RepoOwner:         "cli",
				RepoName:          "cli",
				Asset:             "gh_2.1.0_macOS_arm64.zip",
				Format:            "zip",
				Checksum:          "aaa",
				ChecksumAlgorithm: "sha256",
				Files: []*g2.File{
					{Name: "gh", Src: "gh_2.1.0_macOS_arm64/bin/gh"},
				},
				GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{
					SignerWorkflow2: "cli/cli/.github/workflows/release.yml",
				},
			},
			{
				OS:                "linux",
				Arch:              "amd64",
				Variants:          map[string]string{"libc": "musl"},
				Type:              "github_release",
				RepoOwner:         "cli",
				RepoName:          "cli",
				Asset:             "gh_2.1.0_linux_amd64_musl.tar.gz",
				Format:            "tar.gz",
				Checksum:          "bbb",
				ChecksumAlgorithm: "sha256",
			},
		},
	}
}

func wantPackages() []*lockfile.Package {
	return []*lockfile.Package{
		{
			Name:    "cli/cli",
			Version: "v2.1.0",
			Registry: &lockfile.Registry{
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry-g2",
			},
			OS:                "darwin",
			Arch:              "arm64",
			Type:              "github_release",
			RepoOwner:         "cli",
			RepoName:          "cli",
			Asset:             "gh_2.1.0_macOS_arm64.zip",
			Format:            "zip",
			Checksum:          "aaa",
			ChecksumAlgorithm: "sha256",
			Files: []*lockfile.File{
				{Name: "gh", Src: "gh_2.1.0_macOS_arm64/bin/gh"},
			},
			GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{
				SignerWorkflow2: "cli/cli/.github/workflows/release.yml",
			},
		},
		{
			Name:    "cli/cli",
			Version: "v2.1.0",
			Registry: &lockfile.Registry{
				Type:      "github_content",
				RepoOwner: "aquaproj",
				RepoName:  "aqua-registry-g2",
			},
			OS:                "linux",
			Arch:              "amd64",
			Variants:          map[string]string{"libc": "musl"},
			Type:              "github_release",
			RepoOwner:         "cli",
			RepoName:          "cli",
			Asset:             "gh_2.1.0_linux_amd64_musl.tar.gz",
			Format:            "tar.gz",
			Checksum:          "bbb",
			ChecksumAlgorithm: "sha256",
		},
	}
}

func TestClient_LockPackages(t *testing.T) {
	t.Parallel()
	c := g2.New(nil, nil, "", "")
	if diff := cmp.Diff(wantPackages(), c.LockPackages(testRegistry(), "cli/cli", "v2.1.0")); diff != "" {
		t.Fatalf("LockPackages is wrong (-want +got):\n%s", diff)
	}
}

// A package with no file list must not gain an empty one: the lock file omits files
// it doesn't have, and an empty array would show up as a diff on every rewrite.
func TestClient_LockPackages_noFile(t *testing.T) {
	t.Parallel()
	c := g2.New(nil, nil, "owner", "name")
	pkgs := c.LockPackages(&g2.Registry{
		Assets: []*g2.Asset{{OS: "linux", Arch: "amd64", Type: "go_install"}},
	}, "golang.org/x/tools/gopls", "v0.14.0")
	if len(pkgs) != 1 {
		t.Fatalf("got %d packages, want 1", len(pkgs))
	}
	if pkgs[0].Files != nil {
		t.Errorf("Files is %v, want nil", pkgs[0].Files)
	}
	if pkgs[0].Registry.RepoOwner != "owner" || pkgs[0].Registry.RepoName != "name" {
		t.Errorf("Registry is %+v, want the client's repository", pkgs[0].Registry)
	}
}

// registry.json and a lock entry describe the same thing, so converting one way and
// back has to give what it started with, minus what the branch and the path already
// say.
func TestNewRegistry(t *testing.T) {
	t.Parallel()
	c := g2.New(nil, nil, "", "")
	reg := testRegistry()
	if diff := cmp.Diff(reg, g2.NewRegistry(c.LockPackages(reg, "cli/cli", "v2.1.0"))); diff != "" {
		t.Errorf("the registry is wrong (-want +got):\n%s", diff)
	}
}
