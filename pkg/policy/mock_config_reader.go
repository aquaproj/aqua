package policy

type MockConfigReader struct {
	Cfgs []*Config
	Err  error
}

func (r *MockConfigReader) Read(files []string) ([]*Config, error) {
	return r.Cfgs, r.Err
}
