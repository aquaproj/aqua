---
sidebar_position: 850
---

# `forgejo_release` Package

`aqua >= v2.65.0`

The package is downloaded from the releases of a repository on a [Forgejo](https://forgejo.org/) instance, such as [Codeberg](https://codeberg.org).

Forgejo serves the API it inherited from Gitea, so the instance is asked which versions exist the same way GitHub is: `aqua g` and `aqua up` find the newest release without the registry having to list the versions by hand. A repository on a Gitea instance is [gitea_release](gitea-release-package.md), which is this type under another name, read by the same client.

```yaml
packages:
  - type: forgejo_release
    host: codeberg.org
    repo_owner: mergiraf
    repo_name: mergiraf
    asset: mergiraf_{{.Arch}}-{{.OS}}.{{.Format}}
    format: tar.gz
    description: A syntax-aware git merge driver for a growing collection of programming languages and file formats
    replacements:
      darwin: apple-darwin
      linux: unknown-linux-gnu
      amd64: x86_64
      arm64: aarch64
```

## Required fields

* type
* host: The instance the repository is on, such as `codeberg.org`. GitHub's types need no such field because github.com is the only host they can mean; a Forgejo repository says nothing about which instance holds it
* repo_owner
* repo_name
* asset: The template string of the release asset's name

## Checksums

A checksum file published in the release is read by giving `checksum.type` the same name. It is the instance the package itself is on, so a package of another type is refused rather than asked for from nowhere:

```yaml
    checksum:
      type: forgejo_release
      asset: "{{.Asset}}.sha256"
      algorithm: sha256
```

## Which instances this reaches

`host` is a host name and nothing else -- labels of letters, digits and hyphens, separated by dots. A scheme, a path, userinfo, a port, a dot segment and the characters a pattern is written with are all refused, so `https://codeberg.org`, `example.com/git`, `codeberg.org@elsewhere.example.com`, `forgejo.example.com:3000`, `..` and `codeberg.*` are errors. It names an instance served at the root of its host over HTTPS, which is what a public instance such as codeberg.org or gitea.com is.

The port is refused because the host is a directory as well, under `pkgs/forgejo_release`, and a colon can't be one on Windows. An instance on a port of its own needs a field that says the whole base of it, which is what an instance under a sub-path needs too.

Forgejo can also be installed under a sub-path -- `ROOT_URL = https://example.com/git/` puts the API at `https://example.com/git/api/v1` -- and that is not something a host can say. Such an instance is still an [http](http-package.md) package, as it was before this type existed; what it doesn't get is the version listing.

## The package name

Without `name`, the package is named `<host>/<owner>/<repo>`, e.g. `codeberg.org/mergiraf/mergiraf`. The instance is part of it: codeberg.org's `mergiraf/mergiraf` is not github.com's.

## Verification

[cosign](cosign.md) and [minisign](minisign.md) work, and a signature published beside the asset is named the same way the asset is:

```yaml
    minisign:
      type: forgejo_release
      asset: "{{.Asset}}.minisig"
      public_key: ...
```

Where each of these comes from is checked, including the signatures of the checksum file. A `forgejo_release` file on a package of another type is refused, because such a package is on no instance. A `github_release` file on a `forgejo_release` package is refused when it says no `repo_owner` and `repo_name` of its own: it would take the package's, which on github.com are somebody else's repository. Naming a repository is allowed, so a signature published on a GitHub mirror rather than beside the asset is written as:

```yaml
    minisign:
      type: github_release
      repo_owner: a-mirror
      repo_name: mergiraf
      asset: "{{.Asset}}.minisig"
```

So a [version_override](version-overrides.md) that moves older versions back to `github_release` has to say `checksum`, and any signature, again -- what it doesn't mention it inherits, and what it inherits here is refused.

[slsa_provenance](slsa-provenance.md) and [github_artifact_attestations](github-artifact-attestations.md) are refused, on the checksum file as well as on the package. Both verify what GitHub signed, against a github.com repository, which a Forgejo release doesn't have.

## Private instances

Only what an instance serves to anyone is downloaded: there is nowhere yet to say which credential an instance should be read with, so `private: true` isn't supported.
