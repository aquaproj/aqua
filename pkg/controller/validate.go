package controller

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New() //nolint:gochecknoglobals
