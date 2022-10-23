package aqua

func FilterPackageByTag(pkg *Package, tags, excludedTags map[string]struct{}) bool {
	if len(pkg.Tags) == 0 {
		return len(tags) == 0
	}
	if len(excludedTags) == 0 {
		for _, tag := range pkg.Tags {
			if _, ok := tags[tag]; ok {
				return true
			}
		}
		return false
	}
	for _, tag := range pkg.Tags {
		if _, ok := excludedTags[tag]; ok {
			return false
		}
	}
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
