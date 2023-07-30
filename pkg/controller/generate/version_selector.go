package generate

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
)

type Version struct {
	Name        string
	Version     string
	Description string
	URL         string
}

type VersionSelector interface {
	Find(versions []*Version, hasPreview bool) (int, error)
}

type versionSelector struct{}

func NewVersionSelector() VersionSelector {
	return &versionSelector{}
}

func NewMockVersionSelector(idx int, err error) VersionSelector {
	return &mockVersionSelector{
		idx: idx,
		err: err,
	}
}

type mockVersionSelector struct {
	idx int
	err error
}

func (selector *mockVersionSelector) Find(versions []*Version, hasPreview bool) (int, error) {
	return selector.idx, selector.err
}

func (selector *versionSelector) Find(versions []*Version, hasPreview bool) (int, error) {
	if hasPreview {
		return fuzzyfinder.Find(versions, func(i int) string { //nolint:wrapcheck
			return getVersionItem(versions[i])
		},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i < 0 {
					return "No version matches"
				}
				return getVersionPreview(versions[i], i, w)
			}))
	}
	return fuzzyfinder.Find(versions, func(i int) string { //nolint:wrapcheck
		return getVersionItem(versions[i])
	})
}

func getVersionItem(version *Version) string {
	return version.Version
}

func getVersionPreview(version *Version, i, w int) string {
	if i < 0 {
		return ""
	}
	s := version.Version
	if version.Name != "" && version.Name != version.Version {
		s += fmt.Sprintf(" (%s)", version.Name)
	}
	if version.URL != "" || version.Description != "" {
		s += "\n"
	}
	if version.URL != "" {
		s += fmt.Sprintf("\n%s", version.URL)
	}
	if version.URL != "" {
		s += fmt.Sprintf("\n%s", formatDescription(version.Description, w/2-8)) //nolint:gomnd
	}
	return s
}

func convertStringsToVersions(arr []string) []*Version {
	versions := make([]*Version, len(arr))
	for i, a := range arr {
		versions[i] = &Version{
			Version: a,
		}
	}
	return versions
}
