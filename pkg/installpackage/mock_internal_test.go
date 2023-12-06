package installpackage

import "github.com/spf13/afero"

type MockChecksumCalculator struct {
	Checksum string
	Err      error
}

func (m *MockChecksumCalculator) Calculate(_ afero.Fs, _, _ string) (string, error) {
	return m.Checksum, m.Err
}
