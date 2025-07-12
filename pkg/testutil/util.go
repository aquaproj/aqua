package testutil

import (
	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/spf13/afero"
)

func NewFs(files map[string]string, dirs ...string) (afero.Fs, error) {
	fs := afero.NewMemMapFs()
	for name, body := range files {
		err := afero.WriteFile(fs, name, []byte(body), osfile.FilePermission)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	for _, dir := range dirs {
		err := osfile.MkdirAll(fs, dir)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	return fs, nil
}
