package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
)

func TestGetRootDir(t *testing.T) {
	t.Parallel()
	config.GetRootDir()
}
