---
sidebar_position: 860
---

# `gitea_release` Package

`aqua >= v2.65.0`

The package is downloaded from the releases of a repository on a [Gitea](https://about.gitea.com/) instance, such as [gitea.com](https://gitea.com).

```yaml
packages:
  - type: gitea_release
    host: gitea.com
    repo_owner: gitea
    repo_name: tea
    asset: tea-{{trimV .Version}}-{{.OS}}-{{.Arch}}.xz
    format: xz
    description: A command line tool to interact with Gitea servers
    files:
      - name: tea
        src: tea-{{trimV .Version}}-{{.OS}}-{{.Arch}}
    checksum:
      type: gitea_release
      asset: "{{.Asset}}.sha256"
      algorithm: sha256
```

## Required fields

* type
* host: The instance the repository is on, such as `gitea.com`
* repo_owner
* repo_name
* asset: The template string of the release asset's name

## The same as `forgejo_release`

Everything else about this type is what [forgejo_release](forgejo-release-package.md) says: the package name, what a `host` may be, which checksums and signatures can be read, and what is refused. Forgejo serves the API it inherited from Gitea, so aqua reads both with one client.

The two names are separate so that the day Forgejo and Gitea differ is a change in aqua rather than in every definition that named the other one. A package is on one forge or the other, so a `gitea_release` checksum file or signature belongs to a `gitea_release` package, and the installations are kept apart: the same repository read as a Gitea release is installed under `pkgs/gitea_release/<host>/<owner>/<repo>/`.
