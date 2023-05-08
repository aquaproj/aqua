package testutil

import (
	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/spf13/afero"
)

func NewFs(files map[string]string, dirs ...string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	for name, body := range files {
		if err := afero.WriteFile(fs, name, []byte(body), util.FilePermission); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}
	for _, dir := range dirs {
		if err := util.MkdirAll(fs, dir); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}
	return fs, nil
}
