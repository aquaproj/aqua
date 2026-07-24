package exec_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/checksum"
	"github.com/aquaproj/aqua/v2/pkg/config"
	finder "github.com/aquaproj/aqua/v2/pkg/config-finder"
	reader "github.com/aquaproj/aqua/v2/pkg/config-reader"
	execCtrl "github.com/aquaproj/aqua/v2/pkg/controller/exec"
	"github.com/aquaproj/aqua/v2/pkg/controller/which"
	"github.com/aquaproj/aqua/v2/pkg/cosign"
	"github.com/aquaproj/aqua/v2/pkg/download"
	"github.com/aquaproj/aqua/v2/pkg/ghattestation"
	registry "github.com/aquaproj/aqua/v2/pkg/install-registry"
	"github.com/aquaproj/aqua/v2/pkg/installpackage"
	"github.com/aquaproj/aqua/v2/pkg/link"
	"github.com/aquaproj/aqua/v2/pkg/minisign"
	"github.com/aquaproj/aqua/v2/pkg/osexec"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/policy"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
	"github.com/aquaproj/aqua/v2/pkg/slsa"
	"github.com/aquaproj/aqua/v2/pkg/testutil"
	"github.com/aquaproj/aqua/v2/pkg/unarchive"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func Test_controller_Exec(t *testing.T) { //nolint:funlen
	t.Parallel()
	data := []struct {
		name    string
		files   map[string]string
		dirs    []string
		links   map[string]string
		env     map[string]string
		param   *config.Param
		exeName string
		rt      *runtime.Runtime
		args    []string
		// allowPolicy is the path of a policy file to allow before the command
		// runs, as "aqua policy allow" does.
		allowPolicy string
		isErr       bool
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				CWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			exeName: "aqua-installer",
			dirs: []string{
				"/home/foo/workspace/.git",
			},
			files: map[string]string{
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				"/home/foo/.local/share/aquaproj-aqua/pkgs/github_content/github.com/aquaproj/aqua-installer/v1.0.0/aqua-installer/aqua-installer": "",
				"/home/foo/workspace/aqua-policy.yaml": `
registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- type: local
`,
			},
			allowPolicy: "/home/foo/workspace/aqua-policy.yaml",
		},
		{
			name: "outside aqua",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				CWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/aquaproj-aqua",
				MaxParallelism: 5,
			},
			exeName: "gh",
			env: map[string]string{
				"PATH": "/home/foo/.local/share/aquaproj-aqua/bin:/usr/local/bin:/usr/bin",
			},
			files: map[string]string{
				"/home/foo/workspace/aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`,
				"/home/foo/workspace/registry.yaml": `packages:
- type: github_content
  repo_owner: aquaproj
  repo_name: aqua-installer
  path: aqua-installer
`,
				"/usr/local/foo/gh": "",
			},
			links: map[string]string{
				"../foo/gh": "/usr/local/bin/gh",
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			dir := t.TempDir()
			testutil.WriteFiles(t, dir, d.files, d.dirs...)
			linker := link.New()
			for dest, src := range d.links {
				src = testutil.Abs(dir, src)
				if err := osfile.MkdirAll(filepath.Dir(src)); err != nil {
					t.Fatal(err)
				}
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			testutil.RootParam(dir, d.param)
			env := testutil.RootEnv(dir, d.env)
			policyValidator := policy.NewValidator(d.param)
			if d.allowPolicy != "" {
				if err := policyValidator.Allow(testutil.Abs(dir, d.allowPolicy)); err != nil {
					t.Fatal(err)
				}
			}
			ghDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			osEnv := osenv.NewMock(env)
			whichCtrl := which.New(d.param, finder.NewConfigFinder(), reader.New(d.param), registry.New(d.param, ghDownloader, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), d.rt, osEnv, linker)
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			executor := &osexec.Mock{}
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, linker, nil, &checksum.Calculator{}, unarchive.New(executor), &cosign.MockVerifier{}, &slsa.MockVerifier{}, &minisign.MockVerifier{}, &ghattestation.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{}, vacuum.NewMock(d.param.RootDir, nil, nil))
			policyFinder := policy.NewConfigFinder()
			ctrl := execCtrl.New(pkgInstaller, whichCtrl, executor, osEnv, policy.NewReader(policyValidator, policyFinder, policy.NewConfigReader()), vacuum.NewMock(d.param.RootDir, nil, nil))
			if err := ctrl.Exec(ctx, logger, d.param, d.exeName, d.args...); err != nil {
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

func downloadTestFile(uri, tempDir string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("create a request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send a HTTP request: %w", err)
	}
	defer resp.Body.Close()
	filePath := filepath.Join(tempDir, "registry.yaml")
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create a file: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write a response body to a file: %w", err)
	}
	return filePath, nil
}

// newLocalPolicyReader returns a policy that allows the local registry of the
// benchmark. Without a policy the default one applies, which allows the
// standard registry only, and the package would be rejected before the command
// ever runs.
func newLocalPolicyReader(b *testing.B, dir string) *policy.MockReader {
	b.Helper()
	cfg := &policy.Config{
		Path: filepath.Join(dir, "aqua-policy.yaml"),
		YAML: &policy.ConfigYAML{
			Registries: []*policy.Registry{
				{
					Name: "standard",
					Type: "local",
					Path: filepath.Join(dir, "registry.yaml"),
				},
			},
			Packages: []*policy.Package{
				{
					RegistryName: "standard",
				},
			},
		},
	}
	if err := cfg.Init(); err != nil {
		b.Fatal(err)
	}
	return &policy.MockReader{Configs: []*policy.Config{cfg}}
}

// writeFiles creates the files of a benchmark case in dir.
func writeFiles(b *testing.B, dir string, files map[string]string) {
	b.Helper()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_controller_Exec(b *testing.B) { //nolint:funlen
	data := []struct {
		name    string
		files   map[string]string
		links   map[string]string
		env     map[string]string
		param   *config.Param
		exeName string
		rt      *runtime.Runtime
		args    []string
		isErr   bool
	}{
		{
			name: "normal",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				MaxParallelism: 5,
			},
			exeName: "aqua-installer",
			files:   map[string]string{},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		b.Run("normal", func(b *testing.B) {
			tempDir := b.TempDir()
			ctx := b.Context()
			d.param.CWD = tempDir
			d.param.RootDir = filepath.Join(tempDir, "root")
			d.param.ConfigFilePath = filepath.Join(tempDir, "aqua.yaml")
			d.files["aqua.yaml"] = `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: aquaproj/aqua-installer@v1.0.0
`
			if _, err := downloadTestFile("https://raw.githubusercontent.com/aquaproj/aqua-registry/v2.19.0/registry.yaml", tempDir); err != nil {
				b.Fatal(err)
			}
			writeFiles(b, tempDir, d.files)
			linker := link.New()
			for dest, src := range d.links {
				if err := linker.Symlink(dest, filepath.Join(tempDir, src)); err != nil {
					b.Fatal(err)
				}
			}
			ghDownloader := download.NewGitHubContentFileDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			osEnv := osenv.NewMock(d.env)
			whichCtrl := which.New(d.param, finder.NewConfigFinder(), reader.New(d.param), registry.New(d.param, ghDownloader, d.rt, &cosign.MockVerifier{}, &slsa.MockVerifier{}), d.rt, osEnv, linker)
			downloader := download.NewDownloader(nil, download.NewHTTPDownloader(logger, http.DefaultClient))
			executor := &osexec.Mock{}
			vacuumMock := vacuum.NewMock(d.param.RootDir, nil, nil)
			pkgInstaller := installpackage.New(d.param, downloader, d.rt, linker, nil, &checksum.Calculator{}, unarchive.New(executor), &cosign.MockVerifier{}, &slsa.MockVerifier{}, &minisign.MockVerifier{}, &ghattestation.MockVerifier{}, &installpackage.MockGoInstallInstaller{}, &installpackage.MockGoBuildInstaller{}, &installpackage.MockCargoPackageInstaller{}, vacuumMock)
			ctrl := execCtrl.New(pkgInstaller, whichCtrl, executor, osEnv, newLocalPolicyReader(b, tempDir), vacuumMock)
			b.ResetTimer()
			for b.Loop() {
				func() {
					if err := ctrl.Exec(ctx, logger, d.param, d.exeName, d.args...); err != nil {
						if d.isErr {
							return
						}
						b.Fatal(err)
					}
					if d.isErr {
						b.Fatal("error must be returned")
					}
				}()
			}
		})
	}
}
