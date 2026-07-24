package installpackage

type MockChecksumCalculator struct {
	Checksum string
	Err      error
}

func (m *MockChecksumCalculator) Calculate(filename, algorithm string) (string, error) {
	return m.Checksum, m.Err
}
