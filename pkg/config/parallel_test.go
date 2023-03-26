package config_test

import (
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/sirupsen/logrus"
)

func TestGetMaxParallelism(t *testing.T) {
	t.Parallel()
	data := []struct {
		name              string
		envMaxParallelism string
		exp               int
	}{
		{
			name: "empty",
			exp:  5,
		},
		{
			name:              "invalid",
			envMaxParallelism: "hello",
			exp:               5,
		},
		{
			name:              "10",
			envMaxParallelism: "10",
			exp:               10,
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, d := range data {
		d := d
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			maxParallelism := config.GetMaxParallelism(d.envMaxParallelism, logE)
			if maxParallelism != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, maxParallelism)
			}
		})
	}
}
