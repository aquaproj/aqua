package config

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

// defaultMaxParallelism is the default number of parallel operations when not specified
const defaultMaxParallelism = 5

// GetMaxParallelism determines the maximum number of parallel operations.
// It parses the AQUA_MAX_PARALLELISM environment variable or returns the default value.
func GetMaxParallelism(envMaxParallelism string, logE *logrus.Entry) int {
	if envMaxParallelism == "" {
		return defaultMaxParallelism
	}
	num, err := strconv.Atoi(envMaxParallelism)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"AQUA_MAX_PARALLELISM": envMaxParallelism,
		}).Warn("the environment variable AQUA_MAX_PARALLELISM must be a number")
		return defaultMaxParallelism
	}
	if num <= 0 {
		return defaultMaxParallelism
	}
	return num
}
