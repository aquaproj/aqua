package download

import (
	"errors"
	"fmt"
)

var (
	errInvalidPackageType    = errors.New("package type is invalid")
	errInvalidHTTPStatusCode = errors.New("status code >= 400")
	// errNotFound is a status code of 404. It is an errInvalidHTTPStatusCode too, so
	// anything asking about that still matches.
	errNotFound = fmt.Errorf("%w: not found", errInvalidHTTPStatusCode)
)
