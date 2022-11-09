package aqua

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
