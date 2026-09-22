package g2

import "errors"

var errNoAsset = errors.New("registry.json has no asset")

var errTreeTruncated = errors.New("the versions of the package are too many to list at once")

var errNoVersionOverride = errors.New("registry.yaml has no version_overrides")

// ErrNoPackageBranch is what a package aqua-registry-g2 doesn't hold gets.
//
// It is told apart from any other failure because it means something different: the
// registry has nothing to say about the package, rather than having failed to say
// it. A caller choosing a version can ask upstream instead, since a package with no
// branch can't be locked at any version and narrowing the choice protects nothing.
var ErrNoPackageBranch = errors.New("aqua-registry-g2 doesn't hold the package")
