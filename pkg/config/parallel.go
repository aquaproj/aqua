package config

import (
	"log/slog"
	"strconv"
)

// defaultMaxParallelism is the default number of parallel operations when not specified
const defaultMaxParallelism = 5

// GetMaxParallelism determines the maximum number of parallel operations.
// It parses the AQUA_MAX_PARALLELISM environment variable, then falls back to
// settingMaxParallelism from the settings file, then to the default value.
// An unusable environment variable falls back to the settings file too, because
// it says nothing about how many operations the user wants to run in parallel.
func GetMaxParallelism(envMaxParallelism string, settingMaxParallelism int, logger *slog.Logger) int {
	if envMaxParallelism != "" {
		num, err := strconv.Atoi(envMaxParallelism)
		switch {
		case err != nil:
			logger.Warn("the environment variable AQUA_MAX_PARALLELISM must be a number", "AQUA_MAX_PARALLELISM", envMaxParallelism)
		case num > 0:
			return num
		}
	}
	if settingMaxParallelism > 0 {
		return settingMaxParallelism
	}
	return defaultMaxParallelism
}
