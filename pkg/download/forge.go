package download

import (
	"errors"
	"fmt"
	"net/url"
)

// errForgeHostRequired is returned when nothing says which instance to download from.
// Validation refuses such a package, so this is a file read some other way: a signature
// named as a file on an instance by a package that isn't on one at all.
var errForgeHostRequired = errors.New("the instance to download from is unknown")

// forgeReleaseURL is where a Forgejo or Gitea instance serves one asset of one release.
//
// Such an asset needs no API call to be found: the instance serves it at
// https://<host>/<owner>/<repo>/releases/download/<tag>/<asset>, which is the same
// shape GitHub uses and is what the registry's http packages wrote by hand before
// these types existed.
//
// Each segment is escaped because a tag is upstream's text rather than ours: a
// version such as kustomize/v5.8.1 holds a slash, and left as it is the slash would
// name a path the instance doesn't serve.
func forgeReleaseURL(host, owner, repo, tag, asset string) (string, error) {
	if host == "" {
		return "", errForgeHostRequired
	}
	return fmt.Sprintf("https://%s/%s/%s/releases/download/%s/%s",
		host, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(tag), url.PathEscape(asset)), nil
}
