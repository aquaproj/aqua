package versiongetter

import (
	"regexp"

	"github.com/hashicorp/go-version"
)

var versionPattern = regexp.MustCompile(`^(.*?)v?((?:\d+)(?:\.\d+)?(?:\.\d+)?(?:(\.|-).+)?)$`)

// GetVersionAndPrefix parses a version tag and returns the version and any prefix.
// It attempts to extract a semantic version from tags that may have custom prefixes.
//
// Examples:
//   - "v1.2.3" returns version "1.2.3" with empty prefix
//   - "1.2.3" returns version "1.2.3" with empty prefix
//   - "release-v1.2.3" returns version "1.2.3" with prefix "release-"
//   - "kubernetes-1.28.0" returns version "1.28.0" with prefix "kubernetes-"
//
// Returns nil version and empty prefix if the tag doesn't contain a valid version.
// Returns an error if a version pattern is found but cannot be parsed as a valid semantic version.
func GetVersionAndPrefix(tag string) (*version.Version, string, error) {
	if v, err := version.NewVersion(tag); err == nil {
		return v, "", nil
	}
	a := versionPattern.FindStringSubmatch(tag)
	if a == nil {
		return nil, "", nil
	}
	v, err := version.NewVersion(a[2])
	if err != nil {
		return nil, "", err //nolint:wrapcheck
	}
	return v, a[1], nil
}
