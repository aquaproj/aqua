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
	Find(versions []*Version) (int, error)
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

func (selector *mockVersionSelector) Find(versions []*Version) (int, error) {
	return selector.idx, selector.err
}

func (selector *versionSelector) Find(versions []*Version) (int, error) {
	return fuzzyfinder.Find(versions, func(i int) string { //nolint:wrapcheck
		return getVersionItem(versions[i])
	},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 {
				return "No package matches"
			}
			return getVersionPreview(versions[i], i, w)
		}))
}

func getVersionItem(version *Version) string {
	return version.Version
}

func getVersionPreview(version *Version, i, w int) string {
	if i < 0 {
		return ""
	}
	return fmt.Sprintf("%s (%s)\n\n%s\n%s",
		version.Version,
		version.Name,
		version.URL,
		formatDescription(version.Description, w/2-8)) //nolint:gomnd
}
