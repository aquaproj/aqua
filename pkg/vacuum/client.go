package vacuum

import (
	"io/fs"
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

func (c *Client) dir(pkgPath string) string {
	return filepath.Join(c.rootDir, pkgPath)
}

func (c *Client) file(pkgPath string) string {
	return filepath.Join(c.dir(pkgPath), fileName)
}

func (c *Client) Remove(pkgPath string) error {
	file := c.file(pkgPath)
	if err := c.fs.Remove(file); err != nil {
		return err
	}
	return nil
}

func (c *Client) Update(pkgPath string, timestamp time.Time) error {
	dir := c.dir(pkgPath)
	if err := osfile.MkdirAll(c.fs, dir); err != nil {
		return err
	}
	file := filepath.Join(dir, fileName)
	timestampStr := timestamp.Format(time.RFC3339)
	if err := afero.WriteFile(c.fs, file, []byte(timestampStr), filePermission); err != nil {
		return err
	}
	return nil
}

func (c *Client) FindAll() (map[string]time.Time, error) {
	timestamps := map[string]time.Time{}
	if err := afero.Walk(c.fs, c.rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if name != fileName {
			return nil
		}
		b, err := afero.ReadFile(c.fs, path)
		if err != nil {
			return err
		}
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(b)))
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(c.rootDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		timestamps[rel] = t
		return nil
	}); err != nil {
		return nil, err
	}
	return timestamps, nil
}
