package installpackage

import "github.com/spf13/afero"

type MockChecksumCalculator struct {
	Checksum string
	Err      error
}

func (calc *MockChecksumCalculator) Calculate(fs afero.Fs, filename, algorithm string) (string, error) {
	return calc.Checksum, calc.Err
}
