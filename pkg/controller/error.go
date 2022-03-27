package controller

import "errors"

var (
	errUnknownPkg        = errors.New("unknown package")
	errCommandIsNotFound = errors.New("command is not found")
)
