---
sidebar_position: 850
---

# `forgejo_release` Package

`aqua >= v2.65.0`

The package is downloaded from the releases of a repository on a [Forgejo](https://forgejo.org/) instance, such as [Codeberg](https://codeberg.org).

Forgejo serves the API it inherited from Gitea, so the instance is asked which versions exist the same way GitHub is: `aqua g` and `aqua up` find the newest release without the registry having to list the versions by hand.

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

A checksum file published in the release is read by giving `checksum.type` the same name:

```yaml
    checksum:
      type: forgejo_release
      asset: "{{.Asset}}.sha256"
      algorithm: sha256
```

## Private instances

Only what an instance serves to anyone is downloaded: there is nowhere yet to say which credential an instance should be read with, so `private: true` isn't supported.
