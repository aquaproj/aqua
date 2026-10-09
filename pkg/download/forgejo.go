package download

import (
	"errors"
	"fmt"
	"net/url"
)

// errForgejoHostRequired is returned when nothing says which instance to download from.
// Validation refuses such a package, so this is a file read some other way: a signature
// named as a forgejo_release file by a package that isn't on an instance at all.
var errForgejoHostRequired = errors.New("the instance to download from is unknown")

// forgejoReleaseURL is where a Forgejo instance serves one asset of one release.
//
// A Forgejo release asset needs no API call to be found: the instance serves it at
// https://<host>/<owner>/<repo>/releases/download/<tag>/<asset>, which is the same
// shape GitHub uses and is what the registry's http packages wrote by hand before
// this type existed.
//
// Each segment is escaped because a tag is upstream's text rather than ours: a
// version such as kustomize/v5.8.1 holds a slash, and left as it is the slash would
// name a path the instance doesn't serve.
func forgejoReleaseURL(host, owner, repo, tag, asset string) (string, error) {
	if host == "" {
		return "", errForgejoHostRequired
	}
	return fmt.Sprintf("https://%s/%s/%s/releases/download/%s/%s",
		host, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(tag), url.PathEscape(asset)), nil
}
