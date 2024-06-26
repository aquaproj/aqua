package updateaqua

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) UpdateAqua(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	rootBin := filepath.Join(c.rootDir, "bin")
	if err := osfile.MkdirAll(c.fs, rootBin); err != nil {
		return fmt.Errorf("create the directory: %w", err)
	}

	version, err := c.getVersion(ctx, param)
	if err != nil {
		return err
	}

	logE = logE.WithField("new_version", version)

	if err := c.installer.InstallAqua(ctx, logE, version); err != nil {
		return fmt.Errorf("download aqua: %w", logerr.WithFields(err, logrus.Fields{
			"new_version": version,
		}))
	}
	return nil
}

func (c *Controller) getVersion(ctx context.Context, param *config.Param) (string, error) {
	switch len(param.Args) {
	case 0:
		release, _, err := c.github.GetLatestRelease(ctx, "aquaproj", "aqua")
		if err != nil {
			return "", fmt.Errorf("get the latest version of aqua: %w", err)
		}
		return release.GetTagName(), nil
	case 1:
		return param.Args[0], nil
	default:
		return "", errors.New("too many arguments")
	}
}
