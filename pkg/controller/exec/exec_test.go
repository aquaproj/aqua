package exec_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/clivm/clivm/pkg/config"
	finder "github.com/clivm/clivm/pkg/config-finder"
	reader "github.com/clivm/clivm/pkg/config-reader"
	execCtrl "github.com/clivm/clivm/pkg/controller/exec"
	"github.com/clivm/clivm/pkg/controller/which"
	"github.com/clivm/clivm/pkg/download"
	"github.com/clivm/clivm/pkg/exec"
	registry "github.com/clivm/clivm/pkg/install-registry"
	"github.com/clivm/clivm/pkg/installpackage"
	"github.com/clivm/clivm/pkg/link"
	"github.com/clivm/clivm/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/go-osenv/osenv"
)

func Test_controller_Exec(t *testing.T) { //nolint:funlen
	t.Parallel()
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
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/clivm",
				MaxParallelism: 5,
			},
			exeName: "clivm-installer",
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: clivm
  repo_name: clivm-installer
  path: clivm-installer
`,
				"/home/foo/.local/share/clivm/pkgs/github_content/github.com/clivm/clivm-installer/v1.0.0/clivm-installer/clivm-installer": "",
			},
		},
		{
			name: "outside aqua",
			rt: &runtime.Runtime{
				GOOS:   "linux",
				GOARCH: "amd64",
			},
			param: &config.Param{
				PWD:            "/home/foo/workspace",
				ConfigFilePath: "aqua.yaml",
				RootDir:        "/home/foo/.local/share/clivm",
				MaxParallelism: 5,
			},
			exeName: "gh",
			env: map[string]string{
				"PATH": "/home/foo/.local/share/clivm/bin:/usr/local/bin:/usr/bin",
			},
			files: map[string]string{
				"aqua.yaml": `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`,
				"registry.yaml": `packages:
- type: github_content
  repo_owner: clivm
  repo_name: clivm-installer
  path: clivm-installer
`,
				"/usr/local/foo/gh": "",
			},
			links: map[string]string{
				"../foo/gh": "/usr/local/bin/gh",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			linker := link.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					t.Fatal(err)
				}
			}
			downloader := download.NewRegistryDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			osEnv := osenv.NewMock(d.env)
			whichCtrl := which.New(d.param, finder.NewConfigFinder(fs), reader.New(fs), registry.New(d.param, downloader, fs), d.rt, osEnv, fs, linker)
			pkgDownloader := download.NewPackageDownloader(nil, d.rt, download.NewHTTPDownloader(http.DefaultClient))
			executor := exec.NewMock(0, nil)
			pkgInstaller := installpackage.New(d.param, pkgDownloader, d.rt, fs, linker, executor)
			ctrl := execCtrl.New(pkgInstaller, whichCtrl, executor, osEnv, fs)
			if err := ctrl.Exec(ctx, d.param, d.exeName, d.args, logE); err != nil {
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

func Benchmark_controller_Exec(b *testing.B) { //nolint:funlen,gocognit,cyclop
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
				PWD:            "/home/foo/workspace",
				RootDir:        "/home/foo/.local/share/clivm",
				MaxParallelism: 5,
			},
			exeName: "clivm-installer",
			files:   map[string]string{},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	ctx := context.Background()
	for _, d := range data {
		d := d
		b.Run("normal", func(b *testing.B) {
			tempDir := b.TempDir()
			d.param.ConfigFilePath = filepath.Join(tempDir, "aqua.yaml")
			d.files[d.param.ConfigFilePath] = `registries:
- type: local
  name: standard
  path: registry.yaml
packages:
- name: clivm/clivm-installer@v1.0.0
`
			if _, err := downloadTestFile("https://raw.githubusercontent.com/clivm/clivm-registry/v2.19.0/registry.yaml", tempDir); err != nil {
				b.Fatal(err)
			}
			fs := afero.NewMemMapFs()
			for name, body := range d.files {
				if err := afero.WriteFile(fs, name, []byte(body), 0o644); err != nil {
					b.Fatal(err)
				}
			}
			linker := link.NewMockLinker(fs)
			for dest, src := range d.links {
				if err := linker.Symlink(dest, src); err != nil {
					b.Fatal(err)
				}
			}
			downloader := download.NewRegistryDownloader(nil, download.NewHTTPDownloader(http.DefaultClient))
			osEnv := osenv.NewMock(d.env)
			whichCtrl := which.New(d.param, finder.NewConfigFinder(fs), reader.New(fs), registry.New(d.param, downloader, afero.NewOsFs()), d.rt, osEnv, fs, linker)
			pkgDownloader := download.NewPackageDownloader(nil, d.rt, download.NewHTTPDownloader(http.DefaultClient))
			executor := exec.NewMock(0, nil)
			pkgInstaller := installpackage.New(d.param, pkgDownloader, d.rt, fs, linker, executor)
			ctrl := execCtrl.New(pkgInstaller, whichCtrl, executor, osEnv, fs)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				func() {
					if err := ctrl.Exec(ctx, d.param, d.exeName, d.args, logE); err != nil {
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
