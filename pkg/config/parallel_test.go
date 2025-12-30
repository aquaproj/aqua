package config_test

import (
	"log/slog"
	"testing"

	"github.com/aquaproj/aqua/v2/pkg/config"
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
	logger := slog.New(slog.DiscardHandler)
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			maxParallelism := config.GetMaxParallelism(d.envMaxParallelism, logger)
			if maxParallelism != d.exp {
				t.Fatalf("wanted %v, got %v", d.exp, maxParallelism)
			}
		})
	}
}
