package policy

import (
	"github.com/sirupsen/logrus"
)

type MockReader struct {
	Config  *Config
	Configs []*Config
	Err     error
}

func (r *MockReader) Read(policyFilePaths []string) ([]*Config, error) {
	allowCfgs(r.Configs)
	return r.Configs, r.Err
}

func (r *MockReader) Append(logE *logrus.Entry, aquaYAMLPath string, policies []*Config, globalPolicyPaths map[string]struct{}) ([]*Config, error) {
	return r.Configs, r.Err
}
