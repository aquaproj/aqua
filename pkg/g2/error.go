package g2

import "errors"

var errNoAsset = errors.New("registry.json has no asset")

var errTreeTruncated = errors.New("the versions of the package are too many to list at once")
