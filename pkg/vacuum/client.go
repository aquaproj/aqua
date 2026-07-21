package vacuum

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

const (
	filePermission = 0o644
	fileName       = "timestamp.txt"
	baseDir        = "metadata"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s) //nolint:wrapcheck
}

type Client struct {
	rootDir string
}

func New(param *config.Param) *Client {
	return &Client{
		rootDir: filepath.Join(param.RootDir, baseDir),
	}
}

func (c *Client) Remove(pkgPath string) error {
	file := c.file(pkgPath)
	if err := os.Remove(file); err != nil {
		return fmt.Errorf("reamove a package timestamp file: %w", err)
	}
	return nil
}

func (c *Client) Update(pkgPath string, timestamp time.Time) error {
	dir := c.dir(pkgPath)
	file := filepath.Join(dir, fileName)
	return c.update(file, dir, timestamp)
}

func (c *Client) Create(pkgPath string, timestamp time.Time) error {
	dir := c.dir(pkgPath)
	file := filepath.Join(dir, fileName)
	if osfile.Exists(file) {
		return nil
	}
	return c.update(file, dir, timestamp)
}

func (c *Client) FindAll(logger *slog.Logger) (map[string]time.Time, error) {
	timestamps := map[string]time.Time{}
	if err := filepath.WalkDir(filepath.Join(c.rootDir, "pkgs"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk directory to find timestamp files: %w", err)
		}
		if entry.Name() != fileName {
			return nil
		}
		b, err := os.ReadFile(path) //nolint:gosec // the path comes from walking aqua's own metadata directory
		if err != nil {
			return fmt.Errorf("read a timestamp file: %w", err)
		}
		t, err := ParseTime(strings.TrimSpace(string(b)))
		if err != nil {
			slogerr.WithError(logger, err).Warn("a timestamp file is broken, so recreating it", "timestamp_file", path)
			if err := c.Update(path, time.Now()); err != nil {
				return fmt.Errorf("recreate a broken package timestamp file: %w", err)
			}
			return nil
		}
		rel, err := filepath.Rel(c.rootDir, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("get a relative file path: %w", err)
		}
		timestamps[rel] = t
		return nil
	}); err != nil {
		return nil, fmt.Errorf("find timestamp files: %w", err)
	}
	return timestamps, nil
}

func (c *Client) update(file, dir string, timestamp time.Time) error {
	if err := osfile.MkdirAll(dir); err != nil {
		return fmt.Errorf("create a package metadata directory: %w", err)
	}
	timestampStr := FormatTime(timestamp)
	if err := os.WriteFile(file, []byte(timestampStr+"\n"), filePermission); err != nil {
		return fmt.Errorf("create a package timestamp file: %w", err)
	}
	return nil
}

func (c *Client) dir(pkgPath string) string {
	return filepath.Join(c.rootDir, pkgPath)
}

func (c *Client) file(pkgPath string) string {
	return filepath.Join(c.dir(pkgPath), fileName)
}
