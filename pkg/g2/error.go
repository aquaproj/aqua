package g2

import "errors"

var errNoAsset = errors.New("registry.json has no asset")

var errTreeTruncated = errors.New("the versions of the package are too many to list at once")

var errNoVersionOverride = errors.New("registry.yaml has no version_overrides")
