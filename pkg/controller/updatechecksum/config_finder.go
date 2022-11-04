package updatechecksum

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

type MockConfigFinder struct {
	Files []string
}

func (finder *MockConfigFinder) Finds(wd, configFilePath string) []string {
	return finder.Files
}
