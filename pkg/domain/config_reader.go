package domain

import (
	"github.com/aquaproj/aqua/pkg/policy"
)

type PolicyConfigReader interface {
	Read([]string) ([]*policy.Config, error)
}

type MockPolicyConfigReader struct {
	Cfgs []*policy.Config
	Err  error
}

func (reader *MockPolicyConfigReader) Read(files []string) ([]*policy.Config, error) {
	return reader.Cfgs, reader.Err
}
