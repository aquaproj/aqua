package vacuum

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
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
	fs      afero.Fs
	rootDir string
}

func New(fs afero.Fs, param *config.Param) *Client {
	return &Client{
		fs:      fs,
		rootDir: filepath.Join(param.RootDir, baseDir),
	}
}

func (c *Client) Remove(pkgPath string) error {
	file := c.file(pkgPath)
	if err := c.fs.Remove(file); err != nil {
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
	if f, err := afero.Exists(c.fs, file); err != nil {
		return fmt.Errorf("check whether a package timestamp file exists: %w", err)
	} else if f {
		return nil
	}
	return c.update(file, dir, timestamp)
}

func (c *Client) FindAll(logger *slog.Logger) (map[string]time.Time, error) {
	timestamps := map[string]time.Time{}
	if err := afero.Walk(c.fs, filepath.Join(c.rootDir, "pkgs"), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk directory to find timestamp files: %w", err)
		}
		name := info.Name()
		if name != fileName {
			return nil
		}
		b, err := afero.ReadFile(c.fs, path)
		if err != nil {
			return fmt.Errorf("read a timestamp file: %w", err)
		}
		t, err := ParseTime(strings.TrimSpace(string(b)))
		if err != nil {
			logger.Warn("a timestamp file is broken, so recreating it", slog.Any("error", err), slog.String("timestamp_file", path))
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
	if err := osfile.MkdirAll(c.fs, dir); err != nil {
		return fmt.Errorf("create a package metadata directory: %w", err)
	}
	timestampStr := FormatTime(timestamp)
	if err := afero.WriteFile(c.fs, file, []byte(timestampStr+"\n"), filePermission); err != nil {
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
