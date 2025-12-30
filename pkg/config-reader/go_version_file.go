package reader

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var goVersionPattern = regexp.MustCompile(`(?m)^go (\d+\.\d+.\d+)$`)

func readGoVersionFile(fs afero.Fs, filePath string, pkg *aqua.Package) error {
	p := filepath.Join(filepath.Dir(filePath), pkg.GoVersionFile)
	b, err := afero.ReadFile(fs, p)
	if err != nil {
		return fmt.Errorf("open a go version file: %w", slogerr.With(err,
			slog.String("go_version_file", p),
		))
	}
	matches := goVersionPattern.FindSubmatch(b)
	if len(matches) == 0 {
		return slogerr.With(errors.New("invalid go version file. No go directive is found. The version must be a semver x.y.z"), //nolint:wrapcheck
			slog.String("go_version_file", p),
		)
	}
	pkg.Version = string(matches[1])
	return nil
}
