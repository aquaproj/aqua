package policy

import (
	"errors"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	ErrConfigFileNotFound = errors.New("policy file isn't found")
	errUnAllowedPackage   = slogerr.With(errors.New("this package isn't allowed"),
		"doc", "https://aquaproj.github.io/docs/reference/codes/002",
	)
)
