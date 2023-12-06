package policy

type MockConfigFinder struct {
	path string
	err  error
}

func (f *MockConfigFinder) Find(policyFilePath, wd string) (string, error) {
	return f.path, f.err
}
