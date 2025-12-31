package config

import (
	"log/slog"
	"strconv"
)

// defaultMaxParallelism is the default number of parallel operations when not specified
const defaultMaxParallelism = 5

// GetMaxParallelism determines the maximum number of parallel operations.
// It parses the AQUA_MAX_PARALLELISM environment variable or returns the default value.
func GetMaxParallelism(envMaxParallelism string, logger *slog.Logger) int {
	if envMaxParallelism == "" {
		return defaultMaxParallelism
	}
	num, err := strconv.Atoi(envMaxParallelism)
	if err != nil {
		logger.Warn("the environment variable AQUA_MAX_PARALLELISM must be a number", "AQUA_MAX_PARALLELISM", envMaxParallelism)
		return defaultMaxParallelism
	}
	if num <= 0 {
		return defaultMaxParallelism
	}
	return num
}
