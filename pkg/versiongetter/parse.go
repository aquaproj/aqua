package versiongetter

import (
	"regexp"

	"github.com/hashicorp/go-version"
)

var versionPattern = regexp.MustCompile(`^(.*?)v?((?:\d+)(?:\.\d+)?(?:\.\d+)?(?:(\.|-).+)?)$`)

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
