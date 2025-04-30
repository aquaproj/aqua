package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
)

type Cache struct {
	m       map[string]map[string]*PackageInfo
	fs      afero.Fs
	path    string
	updated bool
}

func NewCache(fs afero.Fs, rootDir, cfgFilePath string) (*Cache, error) {
	c := &Cache{
		m:    map[string]map[string]*PackageInfo{},
		fs:   fs,
		path: filepath.Join(rootDir, "registry-cache", base64.StdEncoding.EncodeToString([]byte(cfgFilePath))+".json"),
	}
	return c, c.read()
}

func (c *Cache) Write() error {
	if !c.updated {
		return nil
	}
	if err := osfile.MkdirAll(c.fs, filepath.Dir(c.path)); err != nil {
		return fmt.Errorf("create a directory: %w", err)
	}
	f, err := c.fs.Create(c.path)
	if err != nil {
		return fmt.Errorf("create a registry cache: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(c.m); err != nil {
		return fmt.Errorf("encode registry cache: %w", err)
	}
	return nil
}

func (c *Cache) Add(rgPath string, pkgInfo *PackageInfo) {
	c.updated = true
	m, ok := c.m[rgPath]
	if !ok {
		m = map[string]*PackageInfo{}
		c.m[rgPath] = m
	}
	m[pkgInfo.GetName()] = pkgInfo
}

func (c *Cache) Get(rgPath, pkgName string) *PackageInfo {
	pkgInfos, ok := c.m[rgPath]
	if !ok {
		return nil
	}
	return pkgInfos[pkgName]
}

func (c *Cache) read() error {
	f, err := c.fs.Open(c.path)
	if err != nil {
		return fmt.Errorf("open a registry cache: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&c.m); err != nil {
		return fmt.Errorf("parse the registry cache file: %w", err)
	}
	return nil
}
