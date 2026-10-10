package download

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

var (
	// errForgeHostRequired is returned when nothing says which instance to download from.
	// Validation refuses such a package, so this is a file read some other way: a
	// signature named as a file on an instance by a package that isn't on one at all.
	errForgeHostRequired = errors.New("the instance to download from is unknown")
	// errForgeTypeUnknown is returned when the type isn't one of the forges, which is a
	// caller asking the wrong question rather than a definition being wrong.
	errForgeTypeUnknown = errors.New("the type is not a release on a forge instance")
)

// forgeReleaseURL is where a forge instance serves one asset of one release.
//
// Such an asset needs no API call to be found, which is what lets an entry say nothing
// but the instance, the repository, the tag and the name. Each forge serves it at a path
// of its own:
//
//	Forgejo and Gitea: https://<host>/<owner>/<repo>/releases/download/<tag>/<asset>
//	GitLab:            https://<host>/<namespace>/<repo>/-/releases/<tag>/downloads/<asset>
//
// The first is the shape GitHub uses, and is what the registry's http packages wrote by
// hand before these types existed. GitLab's is its permanent release link.
//
// Each segment is escaped because a tag is upstream's text rather than ours: a version
// such as kustomize/v5.8.1 holds a slash, and left as it is the slash would name a path
// the instance doesn't serve. A GitLab namespace is the exception: a project in subgroups
// is gitlab-org/security/cli, and those separators are part of where the project is.
func forgeReleaseURL(typ, host, owner, repo, tag, asset string) (string, error) {
	if host == "" {
		return "", errForgeHostRequired
	}
	switch typ {
	case config.PkgInfoTypeForgejoRelease, config.PkgInfoTypeGiteaRelease:
		return fmt.Sprintf("https://%s/%s/%s/releases/download/%s/%s",
			host, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(tag), url.PathEscape(asset)), nil
	case config.PkgInfoTypeGitLabRelease:
		return fmt.Sprintf("https://%s/%s/%s/-/releases/%s/downloads/%s",
			host, escapeNamespace(owner), url.PathEscape(repo), url.PathEscape(tag), url.PathEscape(asset)), nil
	default:
		return "", slogerr.With(errForgeTypeUnknown, "package_type", typ) //nolint:wrapcheck
	}
}

// escapeNamespace escapes each segment of a GitLab namespace, keeping the separators
// between them: the segments say where the project is, and only what is inside one is
// upstream's text.
func escapeNamespace(namespace string) string {
	segments := strings.Split(namespace, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
