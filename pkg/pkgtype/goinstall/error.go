package goinstall

import "errors"

var errGoInstallForbidLatest = errors.New(`the version "latest" is forbidden. Please specify Git tag or commit sha`)
