package reader

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
	"github.com/spf13/afero"
)

var goVersionPattern = regexp.MustCompile(`(?m)^go (\d+\.\d+.\d+)$`)

func readGoVersionFile(fs afero.Fs, filePath string, pkg *aqua.Package) error {
	if pkg.GoVersionFile == "" {
		return nil
	}
	p := filepath.Join(filepath.Dir(filePath), pkg.GoVersionFile)
	b, err := afero.ReadFile(fs, p)
	if err != nil {
		return fmt.Errorf("open a go version file: %w", err)
	}
	matches := goVersionPattern.FindSubmatch(b)
	if len(matches) == 0 {
		return errors.New("invalid go version file. No go directive is found. The version must be a semver x.y.z")
	}
	pkg.Version = string(matches[1])
	return nil
}
