---
sidebar_position: 870
---

# `gitlab_release` Package

`aqua >= v2.65.0`

The package is downloaded from the releases of a project on a [GitLab](https://about.gitlab.com/) instance, such as [gitlab.com](https://gitlab.com).

```yaml
packages:
  - type: gitlab_release
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
* repo_owner: The project's namespace. A project inside subgroups carries all of it, e.g. `gitlab-org/security`
* repo_name
* asset: The template string of the release asset's name

## Optional fields

* host: The instance the project is on. It defaults to `gitlab.com`, which nearly every project on GitLab is on; a self-managed instance says its own host, under the same rule [forgejo_release](forgejo-release-package.md) describes (a host name, so no port and no sub-path)

## Generating the definition

`aqua gr` writes it, the way it does for a GitHub repository, when the package name says the project is on gitlab.com:

```console
$ aqua gr gitlab.com/gitlab-org/cli
```

The name after the host is the project's path, so a project inside subgroups is `gitlab.com/gitlab-org/security/cli`. The asset naming is read off the release's own asset list -- the path each permanent link serves, so a link without one is left out -- and `@<version>` names a release other than the newest.

`-l/--limit` reads that many releases and writes a `version_override` for each way they name their assets, the same as for a GitHub repository. A release dated in the future is left out: its assets aren't published yet, so what it says they are named isn't what they will be.

What it leaves out is what only GitHub answers: a signature identity in a workflow of a github.com repository, and provenance GitHub's verifier checks. A checksum file published with the release is kept, as the release's own.

## Where the asset is downloaded from

GitLab serves a release's asset at a permanent link built from the project, the tag and the asset:

```
https://<host>/<repo_owner>/<repo_name>/-/releases/<version>/downloads/<asset>
```

So nothing is asked of the API to install a version; the API is asked only which versions there are, which is what `aqua g` and `aqua up` need.

`asset` is the path that link serves the file at — the `direct_asset_path` the release link was created with, which is what follows `/downloads/` in the API's `direct_asset_url`. It is not the link's `name`, which is a label beside it: `gitlab-org/gitlab-runner` names one of its links `package: RPM riscv64` and serves it at `packages/rpm/gitlab-runner_riscv64.rpm`. A path may be nested, as that one is.

Two kinds of release link can't be installed this way, and `aqua gr` leaves both out:

- a link created with no `direct_asset_path` has no permanent link at all — GitLab answers with wherever the file was uploaded, such as a generic package's own URL
- a link whose file is on another host, which `gitlab-org/gitlab-runner` is: its assets are on S3, and asking the permanent link for one answers with the page GitLab shows before sending a reader to another site. That is a few hundred bytes of HTML, which a download has no way to tell from an asset

A project publishing only those is an [http](http-package.md) package, naming the file where it is, as it was before this type existed; what it doesn't get is the version listing.

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
