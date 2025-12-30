package policy

import (
	"errors"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	ErrConfigFileNotFound = errors.New("policy file isn't found")
	ErrUnAllowedPackage   = slogerr.With(errors.New("this package isn't allowed"),
		slog.String("doc", "https://aquaproj.github.io/docs/reference/codes/002"),
	)
)
