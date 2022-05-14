package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

func TestGetMaxParallelism(t *testing.T) {
	t.Parallel()
	config.GetMaxParallelism(logrus.NewEntry(logrus.New()))
}
