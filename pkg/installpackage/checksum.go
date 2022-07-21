package installpackage

import "github.com/spf13/afero"

func (inst *installer) ReadChecksumFile(fs afero.Fs, p string) error {
	return inst.checksums.ReadFile(fs, p) //nolint:wrapcheck
}

func (inst *installer) UpdateChecksumFile(fs afero.Fs, p string) error {
	return inst.checksums.UpdateFile(fs, p) //nolint:wrapcheck
}
