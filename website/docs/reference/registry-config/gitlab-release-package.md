---
sidebar_position: 870
---

# `gitlab_release` Package

`aqua >= v2.65.0`

The package is downloaded from the releases of a project on a [GitLab](https://about.gitlab.com/) instance, such as [gitlab.com](https://gitlab.com).

```yaml
packages:
  - type: gitlab_release
    host: gitlab.com
    repo_owner: gitlab-org
    repo_name: cli
    asset: glab_{{trimV .Version}}_{{.OS}}_{{.Arch}}.{{.Format}}
    format: tar.gz
    description: A GitLab CLI tool bringing GitLab to your command line
    files:
      - name: glab
        src: bin/glab
    replacements:
      linux: Linux
      darwin: Darwin
      windows: Windows
```

## Required fields

* type
* host: The instance the project is on, such as `gitlab.com`
* repo_owner: The project's namespace. A project inside subgroups carries all of it, e.g. `gitlab-org/security`
* repo_name
* asset: The template string of the release asset's name

## Where the asset is downloaded from

GitLab serves a release's asset at a permanent link built from the project, the tag and the asset's name:

```
https://<host>/<repo_owner>/<repo_name>/-/releases/<version>/downloads/<asset>
```

So nothing is asked of the API to install a version; the API is asked only which versions there are, which is what `aqua g` and `aqua up` need.

A release whose links were created without a file path has no such permanent link — GitLab then points the asset at wherever it was uploaded, such as a generic package. Such a project is an [http](http-package.md) package, as it was before this type existed, and what it doesn't get is the version listing.

## The same as the other instance types

Everything else is what [forgejo_release](forgejo-release-package.md) says: the package's default name and link (`<host>/<repo_owner>/<repo_name>`), what a `host` may be, which checksums and signatures can be read, and what is refused.

A checksum file published with the release is read by giving `checksum.type` this type's name:

```yaml
    checksum:
      type: gitlab_release
      asset: checksums.txt
      algorithm: sha256
```

[slsa_provenance](slsa-provenance.md) and [github_artifact_attestations](github-artifact-attestations.md) are refused, on the checksum file as well as on the package: both verify what GitHub signed, against a github.com repository. `private` is refused too, because there is nowhere yet to say which credential an instance should be read with.
