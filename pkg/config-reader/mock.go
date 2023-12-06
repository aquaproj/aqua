package reader

import (
	"github.com/aquaproj/aqua/v2/pkg/config/aqua"
)

type MockConfigReader struct {
	Cfg *aqua.Config
	Err error
}

func (r *MockConfigReader) Read(configFilePath string, cfg *aqua.Config) error {
	*cfg = *r.Cfg
	return r.Err
}
