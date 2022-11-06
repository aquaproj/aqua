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
	Read(cfg *policy.Config) error
}

type MockPolicyConfigReader struct {
	Cfg *policy.Config
	Err error
}

func (reader *MockPolicyConfigReader) Read(cfg *policy.Config) error {
	*cfg = *reader.Cfg
	return reader.Err
}
