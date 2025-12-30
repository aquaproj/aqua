package vacuum_test

import (
	"io"
	"log/slog"
	"path"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
)

const rootDir = "/home/foo/.local/share/aquaproj-aqua"

func TestClient_Create(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name       string
		pkgPath    string
		timestamp  string
		files      map[string]string
		expPath    string
		expContent string
	}{
		{
			name:       "create",
			pkgPath:    "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip",
			timestamp:  "2025-01-10T00:15:00+09:00",
			expPath:    path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"),
			expContent: "2025-01-10T00:15:00+09:00\n",
		},
		{
			name:      "file exists",
			pkgPath:   "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip",
			timestamp: "2025-01-10T00:15:00+09:00",
			files: map[string]string{
				path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"): "2025-01-01T00:15:00+09:00\n",
			},
			expPath:    path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"),
			expContent: "2025-01-01T00:15:00+09:00\n",
		},
	}
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for k, v := range tt.files {
				if err := osfile.MkdirAll(fs, filepath.Dir(k)); err != nil {
					t.Fatal(err)
				}
				if err := afero.WriteFile(fs, k, []byte(v), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			client := vacuum.New(fs, &config.Param{
				RootDir: rootDir,
			})
			ts, err := vacuum.ParseTime(tt.timestamp)
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Create(tt.pkgPath, ts); err != nil {
				t.Fatal(err)
			}
			b, err := afero.ReadFile(fs, tt.expPath)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expContent, string(b)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestClient_Update(t *testing.T) { //nolint:dupl
	t.Parallel()
	data := []struct {
		name       string
		pkgPath    string
		timestamp  string
		files      map[string]string
		expPath    string
		expContent string
	}{
		{
			name:       "create",
			pkgPath:    "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip",
			timestamp:  "2025-01-10T00:15:00+09:00",
			expPath:    path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"),
			expContent: "2025-01-10T00:15:00+09:00\n",
		},
		{
			name:      "file exists (overwrite)",
			pkgPath:   "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip",
			timestamp: "2025-01-10T00:15:00+09:00",
			files: map[string]string{
				path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"): "2025-01-01T00:15:00+09:00\n",
			},
			expPath:    path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"),
			expContent: "2025-01-10T00:15:00+09:00\n",
		},
	}
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for k, v := range tt.files {
				if err := osfile.MkdirAll(fs, filepath.Dir(k)); err != nil {
					t.Fatal(err)
				}
				if err := afero.WriteFile(fs, k, []byte(v), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			client := vacuum.New(fs, &config.Param{
				RootDir: rootDir,
			})
			ts, err := vacuum.ParseTime(tt.timestamp)
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Update(tt.pkgPath, ts); err != nil {
				t.Fatal(err)
			}
			b, err := afero.ReadFile(fs, tt.expPath)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expContent, string(b)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestClient_Remove(t *testing.T) {
	t.Parallel()
	data := []struct {
		name    string
		pkgPath string
		files   map[string]string
		expPath string
	}{
		{
			name:    "remove",
			pkgPath: "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip",
			files: map[string]string{
				path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"): "2025-01-01T00:15:00+09:00\n",
			},
			expPath: path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"),
		},
	}
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for k, v := range tt.files {
				if err := osfile.MkdirAll(fs, filepath.Dir(k)); err != nil {
					t.Fatal(err)
				}
				if err := afero.WriteFile(fs, k, []byte(v), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			client := vacuum.New(fs, &config.Param{
				RootDir: rootDir,
			})
			if err := client.Remove(tt.pkgPath); err != nil {
				t.Fatal(err)
			}
			if a, err := afero.Exists(fs, tt.expPath); err != nil {
				t.Fatal(err)
			} else if a {
				t.Fatal("the file still exists")
			}
		})
	}
}

func TestClient_FindAll(t *testing.T) {
	t.Parallel()
	data := []struct {
		name  string
		files map[string]string
		exp   map[string]string
	}{
		{
			name: "normal",
			files: map[string]string{
				path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip", "timestamp.txt"): "2025-01-01T00:15:00+09:00\n",
				path.Join(rootDir, "metadata", "pkgs/github_release/github.com/cli/cli/v2.60.0/gh_2.60.0_macOS_arm64.zip", "timestamp.txt"): "2025-01-20T00:15:00+09:00\n",
			},
			exp: map[string]string{
				"pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip": "2025-01-01T00:15:00+09:00",
				"pkgs/github_release/github.com/cli/cli/v2.60.0/gh_2.60.0_macOS_arm64.zip": "2025-01-20T00:15:00+09:00",
			},
		},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for k, v := range tt.files {
				if err := osfile.MkdirAll(fs, filepath.Dir(k)); err != nil {
					t.Fatal(err)
				}
				if err := afero.WriteFile(fs, k, []byte(v), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			client := vacuum.New(fs, &config.Param{
				RootDir: rootDir,
			})
			timestamps, err := client.FindAll(logger)
			if err != nil {
				t.Fatal(err)
			}
			a := make(map[string]string, len(timestamps))
			for k, v := range timestamps {
				a[k] = vacuum.FormatTime(v)
			}
			if diff := cmp.Diff(tt.exp, a); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
