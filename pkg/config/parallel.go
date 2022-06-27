package config

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

const defaultMaxParallelism = 5

func GetMaxParallelism(envMaxParallelism string, logE *logrus.Entry) int {
	if envMaxParallelism == "" {
		return defaultMaxParallelism
	}
	num, err := strconv.Atoi(envMaxParallelism)
	if err != nil {
		logE.WithFields(logrus.Fields{
			"CLIVM_MAX_PARALLELISM": envMaxParallelism,
		}).Warn("the environment variable CLIVM_MAX_PARALLELISM must be a number")
		return defaultMaxParallelism
	}
	if num <= 0 {
		return defaultMaxParallelism
	}
	return num
}
