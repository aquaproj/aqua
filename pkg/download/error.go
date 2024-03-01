package download

import "errors"

var (
	errInvalidPackageType    = errors.New("package type is invalid")
	errInvalidHTTPStatusCode = errors.New("status code >= 400")
)
