package update

import "regexp"

// github_archive, github_content, go_build, go_install, http
var commitHashPattern = regexp.MustCompile(`\b[0-9a-f]{40}\b`)
