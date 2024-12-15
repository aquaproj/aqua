package reader

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var goVersionPattern = regexp.MustCompile(`(?m)^go (\d+\.\d+.\d+)$`)

func readGoVersionFile(fs afero.Fs, filePath string, pkg *aqua.Package) error {
	p := filepath.Join(filepath.Dir(filePath), pkg.GoVersionFile)
	b, err := afero.ReadFile(fs, p)
	if err != nil {
		return fmt.Errorf("open a go version file: %w", logerr.WithFields(err, logrus.Fields{
			"go_version_file": p,
		}))
	}
	matches := goVersionPattern.FindSubmatch(b)
	if len(matches) == 0 {
		return logerr.WithFields(errors.New("invalid go version file. No go directive is found. The version must be a semver x.y.z"), logrus.Fields{ //nolint:wrapcheck
			"go_version_file": p,
		})
	}
	pkg.Version = string(matches[1])
	return nil
}
