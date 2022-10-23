package aqua

func FilterPackageByTag(pkg *Package, tags map[string]struct{}) bool {
	if len(pkg.Tags) == 0 {
		return len(tags) == 0
	}
	for _, tag := range pkg.Tags {
		if _, ok := tags[tag]; ok {
			return true
		}
	}
	return false
}
