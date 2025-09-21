package aqua

// FilterPackageByTag determines if a package should be included based on tag filtering.
// It checks both included and excluded tags to make the final decision.
func FilterPackageByTag(pkg *Package, tags, excludedTags map[string]struct{}) bool {
	if len(pkg.Tags) == 0 {
		return len(tags) == 0
	}
	if len(excludedTags) == 0 {
		return matchTags(pkg, tags)
	}
	for _, tag := range pkg.Tags {
		if _, ok := excludedTags[tag]; ok {
			return false
		}
	}
	return matchTags(pkg, tags)
}

// matchTags checks if a package matches any of the specified tags.
// Returns true if the package has at least one matching tag or if no tags are specified.
func matchTags(pkg *Package, tags map[string]struct{}) bool {
	if len(tags) == 0 {
		return true
	}
	for _, tag := range pkg.Tags {
		if _, ok := tags[tag]; ok {
			return true
		}
	}
	return false
}
