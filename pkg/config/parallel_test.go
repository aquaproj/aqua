package config_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

func TestGetMaxParallelism(t *testing.T) {
	t.Parallel()
	data := []struct {
		name                  string
		envMaxParallelism     string
		settingMaxParallelism int
		exp                   int
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
		{
			name:                  "settings file",
			settingMaxParallelism: 3,
			exp:                   3,
		},
		{
			name:                  "the environment variable wins",
			envMaxParallelism:     "10",
			settingMaxParallelism: 3,
			exp:                   10,
		},
		{
			name:                  "an invalid environment variable falls back to the settings file",
			envMaxParallelism:     "hello",
			settingMaxParallelism: 3,
			exp:                   3,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			maxParallelism := config.GetMaxParallelism(d.envMaxParallelism, d.settingMaxParallelism, logger)
			if maxParallelism != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, maxParallelism)
			}
		})
	}
}
