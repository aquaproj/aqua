package updatechecksum

type MockConfigFinder struct {
	Files []string
}

func (f *MockConfigFinder) Finds(wd, configFilePath string) []string {
	return f.Files
}
