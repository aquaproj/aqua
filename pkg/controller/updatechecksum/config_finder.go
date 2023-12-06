package updatechecksum

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

type MockConfigFinder struct {
	Files []string
}

func (f *MockConfigFinder) Finds(_, _ string) []string {
	return f.Files
}
