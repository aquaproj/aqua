# Configuration (aqua.yaml)

e.g. [suzuki-shunsuke/my-aqua-config](https://github.com/suzuki-shunsuke/my-aqua-config/blob/main/aqua.yaml)

```yaml
registries:
- type: standard
  ref: v0.10.2 # renovate: depName=suzuki-shunsuke/aqua-registry

packages:
- name: helm/helm@v3.7.0
- name: golangci/golangci-lint@v1.42.1
```

## Configuration File Path

The configuration file path can be specified with the `--config (-c)` option.
If the confgiuration file path isn't specified, the file named `[.]aqua.y[a]ml` would be searched from the current directory to the root directory.
Furthermore, in case of `aqua exec` command the global configuration `~/.aqua/global/[.]aqua.y[a]ml` is also read.

## Environment variables

* `AQUA_LOG_LEVEL`: (default: `info`) Log level
* `AQUA_CONFIG`: configuration file path
* `AQUA_ROOT_DIR`: (default: `$HOME/.aqua`)
* `AQUA_MAX_PARALLELISM`: (default: `5`) The maximum number of packages which are installed in parallel at the same time
* `GITHUB_TOKEN`: GitHub Access Token. This is required to install `github_release` packages

## Configuration attributes

* [registries](#registries): The list of registries
* [packages](#packages): The list of installed packages
* [inline_registry](#inline_registry): The inline registry

## `registries`

e.g.

```yaml
registries:
- type: standard
  ref: v0.10.2 # renovate: depName=suzuki-shunsuke/aqua-registry
```

Registry types

* [standard](#standard-registry): aqua's [Standard Registry](https://github.com/suzuki-shunsuke/aqua-registry)
* [local](#local-registry): local file
* [github_content](#github_content-registry): Get the registry by GitHub Repository Content API

### `standard` registry

e.g.

```yaml
registries:
- type: standard
  ref: v0.10.2 # renovate: depName=suzuki-shunsuke/aqua-registry
```

* `ref`: the Registry Version. Please check [Releases](https://github.com/suzuki-shunsuke/aqua-registry/releases)

This is equivalent to the following definition.

```yaml
registries:
- name: standard
  type: github_content
  repo_owner: suzuki-shunsuke
  repo_name: aqua-registry
  ref: v0.10.2 # renovate: depName=suzuki-shunsuke/aqua-registry
  path: registry.yaml
```

### `local` registry

e.g.

```yaml
registries:
- name: local
  type: local
  path: registry.yaml
```

* `name`: Registry name
* `path`: The file path. Either absolute path or relative path from `aqua.yaml`

Please see [Configuration (registry.yaml)](registry_config.md).

### `github_content` registry

e.g.

```yaml
registries:
- name: foo
  type: github_content
  repo_owner: suzuki-shunsuke
  repo_name: aqua-registry
  ref: v0.10.2 # renovate: depName=suzuki-shunsuke/aqua-registry
  path: registry.yaml
```

* `name`: Registry Name
* `repo_owner`: Repository Owner name
* `repo_name`: Repository name
* `ref`: Repository tag or revision
* `path`: file path from the repository root directory

## `packages`

e.g.

```yaml
packages:
- name: helm/helm
  version: v3.7.0 # renovate: depName=helm/helm
- name: golangci/golangci-lint@v1.42.1
  registry: standard
```

* `name`: (string, required) package name
  * format: `<package name>[@<package version>]`
* `registry`: (string, optional) registry name
  * default value is `standard`
* `version`: (string, optional) package version

The following two configuration is equivalent.

```yaml
- name: golangci/golangci-lint@v1.42.1
  registry: standard
```

```yaml
- name: golangci/golangci-lint
  registry: standard
  version: v1.42.1
```

When you want to update the package with Renovate,
the first style is better because you don't have to write code comments for Renovate's Regex Manager.

```yaml
- name: golangci/golangci-lint
  registry: standard
  version: v1.42.1 # renovate: depName=golangci/golangci-lint
```

If the package name in the code comment is wrong, the package version is changed wrongly.

```yaml
- name: golangci/golangci-lint
  registry: standard
  # depName is wrong!
  version: v1.42.1 # renovate: depName=helm/helm
```

On the other hand, you can prevent such a mis configuration by the first style.

## `inline_registry`

e.g.

```yaml
inline_registry:
  packages:
  - type: github_release
    repo_owner: suzuki-shunsuke
    repo_name: cmdx
    asset: 'cmdx_{{trimV .Version}}_{{.OS}}_{{.Arch}}.tar.gz'
```

Please see [Configuration (registry.yaml)](registry_config.md).
