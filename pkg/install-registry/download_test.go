package registry_test

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/suzuki-shunsuke/flute/flute"
)

const servedRegistry = `packages:
- type: github_release
  repo_owner: suzuki-shunsuke
  repo_name: ci-info
  asset: "ci-info_{{.Arch}}-{{.OS}}.tar.gz"
`

func servingInstaller(t *testing.T, rootDir string) *registry.Installer {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	downloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, &http.Client{
		Transport: &flute.Transport{
			Services: []flute.Service{{
				Endpoint: "https://raw.githubusercontent.com",
				Routes: []flute.Route{{
					Name:     "download a registry",
					Matcher:  &flute.Matcher{Method: "GET", Path: "/corp/registry/v1.0.0/registry.yaml"},
					Response: &flute.Response{Base: http.Response{StatusCode: http.StatusOK}, BodyString: servedRegistry},
				}},
			}},
		},
	}))
	return registry.New(&config.Param{RootDir: rootDir, MaxParallelism: 5}, downloader, &runtime.Runtime{GOOS: "linux", GOARCH: "amd64"}, &cosign.MockVerifier{}, &slsa.MockVerifier{})
}

func corpRegistry() *aqua.Registry {
	return &aqua.Registry{Name: "corp", Type: "github_content", RepoOwner: "corp", RepoName: "registry", Ref: "v1.0.0", Path: "registry.yaml"}
}

// DownloadRegistry downloads the registry even when one is installed, records its
// checksum, and replaces the installed file and the JSON made from it. A stale JSON left
// beside the new file would be what InstallRegistry reads.
func TestInstaller_DownloadRegistry(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	cfgFilePath := filepath.Join(rootDir, "aqua.yaml")
	rgst := corpRegistry()
	p, err := rgst.FilePath(rootDir, cfgFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	writeStale(t, p)

	inst := servingInstaller(t, rootDir)
	chks := checksum.New()
	got, err := inst.DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, cfgFilePath, chks)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.PackageInfos) != 1 || got.PackageInfos[0].RepoName != "ci-info" {
		t.Errorf("got %+v, want the served registry", got.PackageInfos)
	}
	checkRecorded(t, chks, rgst)
	checkInstalled(t, p)
	// InstallRegistry now reads the new content, from the JSON it prefers.
	installed, err := inst.InstallRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, cfgFilePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(installed.PackageInfos) != 1 || installed.PackageInfos[0].RepoName != "ci-info" {
		t.Errorf("InstallRegistry read %+v, want the downloaded registry", installed.PackageInfos)
	}
	assertRegistryFilesArePrivate(t, filepath.Join(rootDir, "registries"))
}

// A file whose checksum isn't the one given is refused, and the installed one is kept.
func TestInstaller_DownloadRegistry_checksumMismatch(t *testing.T) {
	t.Parallel()
	rootDir := t.TempDir()
	cfgFilePath := filepath.Join(rootDir, "aqua.yaml")
	rgst := corpRegistry()
	chks := checksum.New()
	id := checksum.RegistryID(rgst)
	chks.Set(id, &checksum.Checksum{ID: id, Checksum: "0000", Algorithm: "sha256"})
	if _, err := servingInstaller(t, rootDir).DownloadRegistry(t.Context(), slog.New(slog.DiscardHandler), rgst, cfgFilePath, chks); err == nil {
		t.Fatal("an error must be returned")
	}
	p, err := rgst.FilePath(rootDir, cfgFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("a refused file must not be installed")
	}
}

// writeStale installs a registry file and its JSON that don't match what is served.
func writeStale(t *testing.T, p string) {
	t.Helper()
	for path, content := range map[string]string{p: "edited\n", p + ".json": `{"packages": []}`} {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func checkRecorded(t *testing.T, chks *checksum.Checksums, rgst *aqua.Registry) {
	t.Helper()
	want, err := checksum.CalculateReader(strings.NewReader(servedRegistry), "sha256")
	if err != nil {
		t.Fatal(err)
	}
	if c := chks.Get(checksum.RegistryID(rgst)); c == nil || !strings.EqualFold(c.Checksum, want) {
		t.Errorf("recorded %+v, want the checksum of the served file", c)
	}
}

func checkInstalled(t *testing.T, p string) {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != servedRegistry {
		t.Errorf("the installed file is %q", b)
	}
}
