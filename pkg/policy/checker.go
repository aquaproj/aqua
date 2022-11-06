package policy

import (
	"errors"
)

var (
	errUnAllowedPackage  = errors.New("this package isn't allowed")
	errUnAllowedRegistry = errors.New("this registry isn't allowed")
)

type Checker struct{}

func NewChecker() *Checker {
	return &Checker{}
}
