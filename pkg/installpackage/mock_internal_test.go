package installpackage

import "github.com/spf13/afero"

type MockChecksumCalculator struct {
	Checksum string
	Err      error
}

func (m *MockChecksumCalculator) Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	return m.Checksum, m.Err
}
