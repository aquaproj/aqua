//nolint:funlen
package registry_test

import (
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/google/go-cmp/cmp"
	"go.yaml.in/yaml/v3"
)

func TestPackageInfo_GetName(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "foo",
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Name: "foo",
			},
		},
		{
			title: "default",
			exp:   "suzuki-shunsuke/ci-info",
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "forgejo_release",
			exp:   "codeberg.org/mergiraf/mergiraf",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
			},
		},
		{
			title: "gitea_release",
			exp:   "gitea.com/gitea/tea",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				Host:      "gitea.com",
				RepoOwner: "gitea",
				RepoName:  "tea",
			},
		},
		{
			title: "gitlab_release",
			exp:   "gitlab.com/gitlab-org/cli",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org",
				RepoName:  "cli",
			},
		},
		{
			// A GitLab project in subgroups is named by its whole namespace.
			title: "gitlab_release in subgroups",
			exp:   "gitlab.com/gitlab-org/security/cli",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org/security",
				RepoName:  "cli",
			},
		},
		{
			title: "forgejo_release with a name of its own",
			exp:   "mergiraf",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Name:      "mergiraf",
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if name := d.pkgInfo.GetName(); name != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, name)
			}
		})
	}
}

func TestPackageInfo_GetLink(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "http://example.com",
			pkgInfo: &registry.PackageInfo{
				Type: "github_release",
				Link: "http://example.com",
			},
		},
		{
			title: "default",
			exp:   "https://github.com/suzuki-shunsuke/ci-info",
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "forgejo_release",
			exp:   "https://codeberg.org/mergiraf/mergiraf",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
			},
		},
		{
			title: "gitea_release",
			exp:   "https://gitea.com/gitea/tea",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				Host:      "gitea.com",
				RepoOwner: "gitea",
				RepoName:  "tea",
			},
		},
		{
			title: "gitlab_release",
			exp:   "https://gitlab.com/gitlab-org/cli",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org",
				RepoName:  "cli",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if link := d.pkgInfo.GetLink(); link != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, link)
			}
		})
	}
}

func TestPackageInfo_GetFormat(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp:   "tar.gz",
			pkgInfo: &registry.PackageInfo{
				Format: "tar.gz",
			},
		},
		{
			title:   "empty",
			exp:     "",
			pkgInfo: &registry.PackageInfo{},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if format := d.pkgInfo.GetFormat(); format != d.exp {
				t.Fatalf("wanted %s, got %s", d.exp, format)
			}
		})
	}
}

func TestPackageInfo_GetFiles(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		exp     []*registry.File
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			exp: []*registry.File{
				{
					Name: "go",
				},
				{
					Name: "gofmt",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Files: []*registry.File{
					{
						Name: "go",
					},
					{
						Name: "gofmt",
					},
				},
			},
		},
		{
			title: "empty",
			exp: []*registry.File{
				{
					Name: "ci-info",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "has name",
			exp: []*registry.File{
				{
					Name: "cmctl",
				},
			},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "cert-manager",
				RepoName:  "cert-manager",
				Name:      "cert-manager/cert-manager/cmctl",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			files := d.pkgInfo.GetFiles()
			if diff := cmp.Diff(d.exp, files); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPackageInfo_MaybeHasCommand(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		has     []string
		lacks   []string
		pkgInfo *registry.PackageInfo
	}{
		{
			title: "normal",
			has: []string{
				"go", "gofmt", "go.build", "go.override", "go.v", "go.vbuild", "go.voverride",
			},
			lacks: []string{"golang"},
			pkgInfo: &registry.PackageInfo{
				RepoName: "golang",
				Files: []*registry.File{
					{
						Name: "go",
					},
					{
						Name: "gofmt",
					},
				},
				Build: &registry.Build{
					Files: []*registry.File{
						{
							Name: "go.build",
						},
					},
				},
				Overrides: []*registry.Override{
					{
						Files: []*registry.File{
							{
								Name: "go.override",
							},
						},
					},
				},
				VersionOverrides: []*registry.VersionOverride{
					{
						Files: []*registry.File{
							{
								Name: "go.v",
							},
						},
						Build: &registry.Build{
							Files: []*registry.File{
								{
									Name: "go.vbuild",
								},
							},
						},
						Overrides: []*registry.Override{
							{
								Files: []*registry.File{
									{
										Name: "go.voverride",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			title: "empty",
			has:   []string{"ci-info"},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "potentially empty",
			has:   []string{"ci-info", "ci-info.prebuilt"},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Files: []*registry.File{
					{
						Name: "ci-info.prebuilt",
					},
				},
				Build: &registry.Build{},
			},
		},
		{
			title: "case-insensitive",
			has:   []string{"CI-INFO", "Ci-Info.Prebuilt"},
			lacks: []string{"ci-inf"},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Files: []*registry.File{
					{
						Name: "ci-info.prebuilt",
					},
				},
				Build: &registry.Build{},
			},
		},
		{
			title: "has name",
			has:   []string{"cmctl"},
			pkgInfo: &registry.PackageInfo{
				Type:      "github_release",
				RepoOwner: "cert-manager",
				RepoName:  "cert-manager",
				Name:      "cert-manager/cert-manager/cmctl",
			},
		},
	}

	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			for _, exe := range d.has {
				if !d.pkgInfo.MaybeHasCommand(exe) {
					t.Fatalf("expected to have command %s", exe)
				}
			}
			for _, exe := range d.lacks {
				if d.pkgInfo.MaybeHasCommand(exe) {
					t.Fatalf("expected to not have command %s", exe)
				}
			}
		})
	}
}

func TestPackageInfo_Validate(t *testing.T) {
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title:   "package name is required",
			pkgInfo: &registry.PackageInfo{},
			isErr:   true,
		},
		{
			title: "repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubArchive,
			},
			isErr: true,
		},
		{
			title: "github_archive",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubArchive,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
		},
		{
			title: "github_content repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubContent,
			},
			isErr: true,
		},
		{
			title: "github_content path is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubContent,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
			isErr: true,
		},
		{
			title: "github_content",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubContent,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Path:      "bin/ci-info",
			},
		},
		{
			title: "github_release repo is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeGitHubRelease,
			},
			isErr: true,
		},
		{
			title: "github_release asset is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
			},
			isErr: true,
		},
		{
			title: "github_release",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     "ci-info.tar.gz",
			},
		},
		{
			title: "http url is required",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeHTTP,
			},
			isErr: true,
		},
		{
			title: "http",
			pkgInfo: &registry.PackageInfo{
				Type: registry.PkgInfoTypeHTTP,
				Name: "suzuki-shunsuke/ci-info",
				URL:  "http://example.com",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
				if d.isErr {
					return
				}
				t.Fatal(err)
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

// What aqua remove is given is a registry's packages, which nothing validated: whatever a
// definition says the instance is, the path has to be one under where the type's packages
// go, and one path rather than a pattern matching several.
func TestPackageInfo_PkgPaths_forgeRelease(t *testing.T) {
	t.Parallel()
	data := []struct {
		title string
		typ   string
		host  string
		exp   string
	}{
		{
			title: "a Forgejo instance",
			typ:   registry.PkgInfoTypeForgejoRelease,
			host:  "codeberg.org",
			exp:   filepath.Join("forgejo_release", "codeberg.org", "mergiraf", "mergiraf"),
		},
		{
			// Each type's packages are its own: the same repository read as a Gitea
			// release is not the same installation.
			title: "a Gitea instance",
			typ:   registry.PkgInfoTypeGiteaRelease,
			host:  "gitea.com",
			exp:   filepath.Join("gitea_release", "gitea.com", "mergiraf", "mergiraf"),
		},
		{
			title: "a GitLab instance",
			typ:   registry.PkgInfoTypeGitLabRelease,
			host:  "gitlab.com",
			exp:   filepath.Join("gitlab_release", "gitlab.com", "mergiraf", "mergiraf"),
		},
		{
			title: "a host that is a path out of the packages",
			typ:   registry.PkgInfoTypeForgejoRelease,
			host:  "../../..",
		},
		{
			// aqua remove expands the path as a glob, so this one would be every
			// instance's copy of the same owner and name.
			title: "a host that is a pattern",
			typ:   registry.PkgInfoTypeForgejoRelease,
			host:  "*",
		},
		{
			title: "a host that is a dot",
			typ:   registry.PkgInfoTypeForgejoRelease,
			host:  ".",
		},
		{
			title: "no instance at all",
			typ:   registry.PkgInfoTypeForgejoRelease,
			host:  "",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			pkgInfo := &registry.PackageInfo{
				Type:      d.typ,
				Host:      d.host,
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
			}
			paths := pkgInfo.PkgPaths()
			if d.exp == "" {
				if len(paths) != 0 {
					t.Fatalf("wanted no path, got %v", paths)
				}
				return
			}
			if _, ok := paths[d.exp]; !ok || len(paths) != 1 {
				t.Fatalf("wanted %s, got %v", d.exp, paths)
			}
		})
	}
}

// The host is what says which instance the release is on, and it is read as a directory as
// well, so what it may be is narrow.
func TestPackageInfo_Validate_forgeRelease_host(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title: "forgejo_release host is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
			},
			isErr: true,
		},
		{
			title: "forgejo_release host written as a URL",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "https://codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
			},
			isErr: true,
		},
		{
			title: "forgejo_release host holding a path",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "example.com/git",
				RepoOwner: "an-owner",
				RepoName:  "a-repo",
				Asset:     "a-repo.tar.gz",
			},
			isErr: true,
		},
		{
			title: "forgejo_release host carrying userinfo",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org@evil.example.com",
				RepoOwner: "an-owner",
				RepoName:  "a-repo",
				Asset:     "a-repo.tar.gz",
			},
			isErr: true,
		},
		{
			title: "forgejo_release host that is a pattern",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.*",
				RepoOwner: "an-owner",
				RepoName:  "a-repo",
				Asset:     "a-repo.tar.gz",
			},
			isErr: true,
		},
		{
			title: "forgejo_release host holding a character a host name can't",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org,elsewhere.example.com",
				RepoOwner: "an-owner",
				RepoName:  "a-repo",
				Asset:     "a-repo.tar.gz",
			},
			isErr: true,
		},
		{
			title: "forgejo_release host that is a dot segment",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "..",
				RepoOwner: "an-owner",
				RepoName:  "a-repo",
				Asset:     "a-repo.tar.gz",
			},
			isErr: true,
		},
		{
			// The host is a directory as well, and a colon can't be one on Windows.
			title: "forgejo_release host with a port",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "forgejo.example.com:3000",
				RepoOwner: "an-owner",
				RepoName:  "a-repo",
				Asset:     "a-repo.tar.gz",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
				if !d.isErr {
					t.Fatal(err)
				}
				return
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

// What else in a definition is a thing only github.com answers, and where the files beside
// the asset come from.
func TestPackageInfo_Validate_forgeRelease(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title: "forgejo_release repo is required",
			pkgInfo: &registry.PackageInfo{
				Type:  registry.PkgInfoTypeForgejoRelease,
				Name:  "codeberg.org/mergiraf/mergiraf",
				Host:  "codeberg.org",
				Asset: "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
			},
			isErr: true,
		},
		{
			title: "forgejo_release asset is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
			},
			isErr: true,
		},
		{
			title: "forgejo_release private is not supported",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
				Private:   true,
			},
			isErr: true,
		},
		{
			title: "a github_release signature on a forgejo_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Minisign: &registry.Minisign{
					Type: registry.PkgInfoTypeGitHubRelease,
				},
			},
			isErr: true,
		},
		{
			title: "a forgejo_release signature on a github_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     "ci-info.tar.gz",
				Cosign: &registry.Cosign{
					Signature: &registry.DownloadedFile{
						Type: registry.PkgInfoTypeForgejoRelease,
					},
				},
			},
			isErr: true,
		},
		{
			title: "a github_release signature of the checksum file, inheriting the repository",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeForgejoRelease,
					Asset: "{{.Asset}}.sha256",
					Minisign: &registry.Minisign{
						Type: registry.PkgInfoTypeGitHubRelease,
					},
				},
			},
			isErr: true,
		},
		{
			title: "a github_release cosign signature of the checksum file, inheriting the repository",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeForgejoRelease,
					Asset: "{{.Asset}}.sha256",
					Cosign: &registry.Cosign{
						Signature: &registry.DownloadedFile{
							Type: registry.PkgInfoTypeGitHubRelease,
						},
					},
				},
			},
			isErr: true,
		},
		{
			title: "attestations of the checksum file of a forgejo_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Checksum: &registry.Checksum{
					Type:                       registry.PkgInfoTypeForgejoRelease,
					Asset:                      "{{.Asset}}.sha256",
					GitHubArtifactAttestations: &registry.GitHubArtifactAttestations{},
				},
			},
			isErr: true,
		},
		{
			// Wherever else a signature is published is nobody's mistake: what the
			// check is for is a file that would take the package's owner and name,
			// which on github.com are somebody else's repository.
			title: "a github_release signature naming the repository it is in",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Minisign: &registry.Minisign{
					Type:      registry.PkgInfoTypeGitHubRelease,
					RepoOwner: "a-mirror",
					RepoName:  "mergiraf",
				},
			},
		},
		{
			title: "gitea_release",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				Host:      "gitea.com",
				RepoOwner: "gitea",
				RepoName:  "tea",
				Asset:     "tea-{{trimV .Version}}-{{.OS}}-{{.Arch}}",
			},
		},
		{
			title: "gitea_release host is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				RepoOwner: "gitea",
				RepoName:  "tea",
				Asset:     "tea",
			},
			isErr: true,
		},
		{
			title: "gitea_release private is not supported",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				Host:      "gitea.com",
				RepoOwner: "gitea",
				RepoName:  "tea",
				Asset:     "tea",
				Private:   true,
			},
			isErr: true,
		},
		{
			// One client reads both, but a definition says which forge it is on, and
			// a file beside the asset is on the same one.
			title: "a gitea_release checksum file on a forgejo_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeGiteaRelease,
					Asset: "{{.Asset}}.sha256",
				},
			},
			isErr: true,
		},
		{
			title: "a gitea_release checksum file on a gitea_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				Host:      "gitea.com",
				RepoOwner: "gitea",
				RepoName:  "tea",
				Asset:     "tea.xz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeGiteaRelease,
					Asset: "{{.Asset}}.sha256",
				},
			},
		},
		{
			title: "a forgejo_release checksum file on a forgejo_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeForgejoRelease,
					Asset: "{{.Asset}}.sha256",
				},
			},
		},
		{
			title: "a forgejo_release checksum file on a github_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitHubRelease,
				RepoOwner: "suzuki-shunsuke",
				RepoName:  "ci-info",
				Asset:     "ci-info.tar.gz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeForgejoRelease,
					Asset: "{{.Asset}}.sha256",
				},
			},
			isErr: true,
		},
		{
			title: "forgejo_release doesn't support what only GitHub can verify",
			pkgInfo: &registry.PackageInfo{
				Type:           registry.PkgInfoTypeForgejoRelease,
				Host:           "codeberg.org",
				RepoOwner:      "mergiraf",
				RepoName:       "mergiraf",
				Asset:          "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
				SLSAProvenance: &registry.SLSAProvenance{},
			},
			isErr: true,
		},
		{
			title: "forgejo_release",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf",
				RepoName:  "mergiraf",
				Asset:     "mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
				if !d.isErr {
					t.Fatal(err)
				}
				return
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}

func TestPackageInfo_YAMLDecode_NestedStructure(t *testing.T) { //nolint:cyclop
	t.Parallel()

	// Test more complex YAML structure with multiple verification configurations
	yamlData := `
name: complex-package
type: github_release
repo_owner: owner
repo_name: repo
asset: package.tar.gz
version_overrides:
  - version_constraint: ">=v1.0.0"
    asset: package-v1.tar.gz
    cosign:
      enabled: true
    minisign:
      enabled: false
  - version_constraint: ">=v2.0.0"
    asset: package-v2.tar.gz
    slsa_provenance:
      enabled: true
      type: github_release
`

	var pkgInfo registry.PackageInfo
	if err := yaml.Unmarshal([]byte(yamlData), &pkgInfo); err != nil {
		t.Fatalf("failed to unmarshal complex YAML: %v", err)
	}

	// Verify version overrides
	if len(pkgInfo.VersionOverrides) != 2 {
		t.Fatalf("expected 2 version overrides, got %d", len(pkgInfo.VersionOverrides))
	}

	// Test first version override
	vo1 := pkgInfo.VersionOverrides[0]
	if vo1.Asset != "package-v1.tar.gz" {
		t.Errorf("expected asset 'package-v1.tar.gz', got %q", vo1.Asset)
	}

	if vo1.Cosign == nil {
		t.Fatal("Cosign should not be nil in first version override")
	}

	if !vo1.Cosign.GetEnabled() {
		t.Error("Cosign should be enabled in first version override")
	}

	if vo1.Minisign == nil {
		t.Fatal("Minisign should not be nil in first version override")
	}

	if vo1.Minisign.GetEnabled() {
		t.Error("Minisign should be disabled in first version override")
	}

	// Test second version override
	vo2 := pkgInfo.VersionOverrides[1]
	if vo2.Asset != "package-v2.tar.gz" {
		t.Errorf("expected asset 'package-v2.tar.gz', got %q", vo2.Asset)
	}

	if vo2.SLSAProvenance == nil {
		t.Fatal("SLSAProvenance should not be nil in second version override")
	}

	if !vo2.SLSAProvenance.GetEnabled() {
		t.Error("SLSAProvenance should be enabled in second version override")
	}
}

// GitLab is read as a package on a forge instance like the other two, and is the first
// type where the owner holding separators of its own is ordinary.
func TestPackageInfo_Validate_gitLabRelease(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		title   string
		pkgInfo *registry.PackageInfo
		isErr   bool
	}{
		{
			title: "gitlab_release",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org",
				RepoName:  "cli",
				Asset:     "glab_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}",
			},
		},
		{
			// A project in subgroups: its namespace is where the project is, and
			// every segment of it is a name a directory can have.
			title: "a project in subgroups",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org/security",
				RepoName:  "cli",
				Asset:     "glab.tar.gz",
			},
		},
		{
			title: "gitlab_release host is required",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				RepoOwner: "gitlab-org",
				RepoName:  "cli",
				Asset:     "glab.tar.gz",
			},
			isErr: true,
		},
		{
			title: "gitlab_release private is not supported",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org",
				RepoName:  "cli",
				Asset:     "glab.tar.gz",
				Private:   true,
			},
			isErr: true,
		},
		{
			title: "gitlab_release doesn't support what only GitHub can verify",
			pkgInfo: &registry.PackageInfo{
				Type:           registry.PkgInfoTypeGitLabRelease,
				Host:           "gitlab.com",
				RepoOwner:      "gitlab-org",
				RepoName:       "cli",
				Asset:          "glab.tar.gz",
				SLSAProvenance: &registry.SLSAProvenance{},
			},
			isErr: true,
		},
		{
			// The owner is a directory as well, and a dot segment would be a path out
			// of where the type's packages go.
			title: "an owner that is a path out of the packages",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org/..",
				RepoName:  "cli",
				Asset:     "glab.tar.gz",
			},
			isErr: true,
		},
		{
			// aqua remove expands the install path as a glob.
			title: "a name that is a pattern",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org",
				RepoName:  "*",
				Asset:     "glab.tar.gz",
			},
			isErr: true,
		},
		{
			// A slash in the owner is GitLab's namespace and nothing else's: on the
			// other instances it would be a path where a name belongs.
			title: "a forgejo_release owner holding a separator",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeForgejoRelease,
				Host:      "codeberg.org",
				RepoOwner: "mergiraf/elsewhere",
				RepoName:  "mergiraf",
				Asset:     "mergiraf.tar.gz",
			},
			isErr: true,
		},
		{
			title: "a gitlab_release checksum file on a gitea_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGiteaRelease,
				Host:      "gitea.com",
				RepoOwner: "gitea",
				RepoName:  "tea",
				Asset:     "tea.xz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeGitLabRelease,
					Asset: "{{.Asset}}.sha256",
				},
			},
			isErr: true,
		},
		{
			title: "a gitlab_release checksum file on a gitlab_release package",
			pkgInfo: &registry.PackageInfo{
				Type:      registry.PkgInfoTypeGitLabRelease,
				Host:      "gitlab.com",
				RepoOwner: "gitlab-org",
				RepoName:  "cli",
				Asset:     "glab.tar.gz",
				Checksum: &registry.Checksum{
					Type:  registry.PkgInfoTypeGitLabRelease,
					Asset: "checksums.txt",
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			if err := d.pkgInfo.Validate(); err != nil {
				if !d.isErr {
					t.Fatal(err)
				}
				return
			}
			if d.isErr {
				t.Fatal("error must be returned")
			}
		})
	}
}
