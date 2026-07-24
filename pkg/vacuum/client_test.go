package vacuum_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/google/go-cmp/cmp"
)

const pkgPath = "pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip"

// newRootDir creates a root directory and the given files in it.
// The keys of files are paths relative to the root directory.
func newRootDir(t *testing.T, files map[string]string) string {
	t.Helper()
	rootDir := t.TempDir()
	for k, v := range files {
		p := filepath.Join(rootDir, filepath.FromSlash(k))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(v), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return rootDir
}

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
			pkgPath:    pkgPath,
			timestamp:  "2025-01-10T00:15:00+09:00",
			expPath:    "metadata/" + pkgPath + "/timestamp.txt",
			expContent: "2025-01-10T00:15:00+09:00\n",
		},
		{
			name:      "file exists",
			pkgPath:   pkgPath,
			timestamp: "2025-01-10T00:15:00+09:00",
			files: map[string]string{
				"metadata/" + pkgPath + "/timestamp.txt": "2025-01-01T00:15:00+09:00\n",
			},
			expPath:    "metadata/" + pkgPath + "/timestamp.txt",
			expContent: "2025-01-01T00:15:00+09:00\n",
		},
	}
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rootDir := newRootDir(t, tt.files)
			client := vacuum.New(&config.Param{
				RootDir: rootDir,
			})
			ts, err := vacuum.ParseTime(tt.timestamp)
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Create(tt.pkgPath, ts); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(tt.expPath)))
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
			pkgPath:    pkgPath,
			timestamp:  "2025-01-10T00:15:00+09:00",
			expPath:    "metadata/" + pkgPath + "/timestamp.txt",
			expContent: "2025-01-10T00:15:00+09:00\n",
		},
		{
			name:      "file exists (overwrite)",
			pkgPath:   pkgPath,
			timestamp: "2025-01-10T00:15:00+09:00",
			files: map[string]string{
				"metadata/" + pkgPath + "/timestamp.txt": "2025-01-01T00:15:00+09:00\n",
			},
			expPath:    "metadata/" + pkgPath + "/timestamp.txt",
			expContent: "2025-01-10T00:15:00+09:00\n",
		},
	}
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rootDir := newRootDir(t, tt.files)
			client := vacuum.New(&config.Param{
				RootDir: rootDir,
			})
			ts, err := vacuum.ParseTime(tt.timestamp)
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Update(tt.pkgPath, ts); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(rootDir, filepath.FromSlash(tt.expPath)))
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
			pkgPath: pkgPath,
			files: map[string]string{
				"metadata/" + pkgPath + "/timestamp.txt": "2025-01-01T00:15:00+09:00\n",
			},
			expPath: "metadata/" + pkgPath + "/timestamp.txt",
		},
	}
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rootDir := newRootDir(t, tt.files)
			client := vacuum.New(&config.Param{
				RootDir: rootDir,
			})
			if err := client.Remove(tt.pkgPath); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(rootDir, filepath.FromSlash(tt.expPath))); !os.IsNotExist(err) {
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
				"metadata/pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip/timestamp.txt": "2025-01-01T00:15:00+09:00\n",
				"metadata/pkgs/github_release/github.com/cli/cli/v2.60.0/gh_2.60.0_macOS_arm64.zip/timestamp.txt": "2025-01-20T00:15:00+09:00\n",
			},
			exp: map[string]string{
				"pkgs/github_release/github.com/cli/cli/v2.65.0/gh_2.65.0_macOS_arm64.zip": "2025-01-01T00:15:00+09:00",
				"pkgs/github_release/github.com/cli/cli/v2.60.0/gh_2.60.0_macOS_arm64.zip": "2025-01-20T00:15:00+09:00",
			},
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, tt := range data {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rootDir := newRootDir(t, tt.files)
			client := vacuum.New(&config.Param{
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
