package domain

import (
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/policy"
)

type ConfigReader interface {
	Read(configFilePath string, cfg *aqua.Config) error
}

type MockConfigReader struct {
	Cfg *aqua.Config
	Err error
}

func (reader *MockConfigReader) Read(configFilePath string, cfg *aqua.Config) error {
	*cfg = *reader.Cfg
	return reader.Err
}

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
