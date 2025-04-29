package registry

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Config struct {
	PackageInfos PackageInfos `yaml:"packages" json:"packages"`
	m            map[string]*PackageInfo
}

type Cache struct {
	m       map[string]map[string]*PackageInfo
	fs      afero.Fs
	path    string
	updated bool
}

func NewCache(fs afero.Fs, p string) *Cache {
	return &Cache{
		m:    map[string]map[string]*PackageInfo{},
		fs:   fs,
		path: p,
	}
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
	pkgInfo, _ := pkgInfos[pkgName]
	return pkgInfo
}

func (c *Config) Packages(logE *logrus.Entry) map[string]*PackageInfo {
	if c.m != nil {
		return c.m
	}
	m := c.PackageInfos.ToMap(logE)
	c.m = m
	return m
}

func (c *Config) Package(logE *logrus.Entry, pkgName string) *PackageInfo {
	m := c.Packages(logE)
	pkg, ok := m[pkgName]
	if !ok {
		return nil
	}
	return pkg
}
