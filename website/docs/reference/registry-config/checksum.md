---
sidebar_position: 1950
---

# checksum

Each registry's package configuration has the configuration about checksum.

e.g. [GitHub CLI](https://github.com/aquaproj/aqua-registry/blob/109811850abf8ec34f8715f3384ba8218f05ec1d/pkgs/cli/cli/registry.yaml)

```yaml
packages:
  - type: github_release
    repo_owner: cli
    repo_name: cli
    # ...
    checksum:
      type: github_release
      asset: gh_{{trimV .Version}}_checksums.txt
      file_format: regexp
      algorithm: sha256
      pattern:
        checksum: ^(\b[A-Fa-f0-9]{64}\b)
        file: "^\\b[A-Fa-f0-9]{64}\\b\\s+(\\S+)$"
```

e.g. [ArgoCD CLI](https://github.com/aquaproj/aqua-registry/blob/109811850abf8ec34f8715f3384ba8218f05ec1d/pkgs/argoproj/argo-cd/registry.yaml)

```yaml
packages:
  - type: github_release
    repo_owner: argoproj
    repo_name: argo-cd
    # ...
    asset: argocd-{{.OS}}-{{.Arch}}
    checksum:
      type: github_release
      asset: "{{.Asset}}.sha256"
      file_format: raw
      algorithm: sha256
```

## `enabled`

You can enable or disable the checksum download by `enabled` attribute.

```yaml
packages:
  - type: github_release
    repo_owner: argoproj
    repo_name: argo-cd
    # ...
    checksum:
      enabled: false
```

By default, checksum download is disabled.

## checksum `algorithm`

The following `algorithm` are supported.

* `sha1` (aqua >= [v1.29.0](https://github.com/aquaproj/aqua/releases/tag/v1.29.0))
* `sha256`
* `sha512`
* `md5`

## checksum `type`

The following `type` are supported.

- `github_release`
- `http`

`github_release` requires the following attributes.

- `asset`: GitHub Release Asset name. The format is a Go's `text/template` string

`http` requires the following attributes.

- `url`: Checksum file's download URL. The format is a Go's `text/template` string

## checksum `file_format`

The following `file_format` are supported.

- `regexp`
- `raw`

`regexp` requires the following attributes.

- `pattern.checksum`:
- `pattern.file`:

```yaml
pattern:
  checksum: ^(\b[A-Fa-f0-9]{64}\b)
  file: "^\\b[A-Fa-f0-9]{64}\\b\\s+(\\S+)$"
```

aqua extracts pairs of checksum and asset name using regular expressions.
If the checksum file includes only one checksum, you can omit `pattern.file`.
