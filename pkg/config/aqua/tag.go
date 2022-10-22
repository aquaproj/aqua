package aqua

func FilterPackageByTag(pkg *Package, tags map[string]struct{}) bool {
	if len(tags) == 0 {
		return true
	}
	if len(pkg.Tags) == 0 {
		return true
	}
	for _, tag := range pkg.Tags {
		if _, ok := tags[tag]; ok {
			return true
		}
	}
	return false
}

func FilterPackagesByTag(pkgs []*Package, tags map[string]struct{}) []*Package {
	if len(tags) == 0 {
		return pkgs
	}
	arr := make([]*Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		pkg := pkg
		if FilterPackageByTag(pkg, tags) {
			arr = append(arr, pkg)
		}
	}
	return arr
}
