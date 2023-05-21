package cargo

import (
	"errors"
)

var errHTTPStatusCodeIsGreaterEqualThan300 = errors.New("HTTP status code is greater equal than 300")
