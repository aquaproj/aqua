package generate

import (
	"github.com/aquaproj/aqua/v2/pkg/github"
)

func filterTag(tag *github.RepositoryTag, filters []*Filter) bool {
	tagName := tag.GetName()
	for _, filter := range filters {
		if filterTagByFilter(tagName, filter) {
			return true
		}
	}
	return false
}
