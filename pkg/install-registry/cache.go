package registry

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (is *Installer) cachePath(cfgFilePath string) string {
	return filepath.Join(is.param.RootDir, "registry-cache", base64.StdEncoding.EncodeToString([]byte(cfgFilePath))+".json")
}

func (is *Installer) ReadCache(logE *logrus.Entry, cfgFilePath string) (*registry.Cache, error) {
	cachePath := is.cachePath(cfgFilePath)
	m := registry.NewCache(is.fs, cachePath)
	f, err := is.fs.Open(cachePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"cache_path": cachePath,
			}).Warn("open a registry cache file. Will remove and recreate it")
			if err := is.fs.Remove(cachePath); err != nil {
				logerr.WithError(logE, err).WithFields(logrus.Fields{
					"cache_path": cachePath,
				}).Warn("remove a registry cache file")
			} else {
				logE.WithFields(logrus.Fields{
					"cache_path": cachePath,
				}).Debug("remove a registry cache file")
			}
			return nil, nil
		}
		return nil, nil
	}
	if f != nil {
		defer f.Close()
	}
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, fmt.Errorf("parse the registry cache file: %w", err)
	}

	return m, nil
}
