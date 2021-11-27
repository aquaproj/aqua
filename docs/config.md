# Configuration (aqua.yaml)

e.g. [suzuki-shunsuke/my-aqua-config](https://github.com/suzuki-shunsuke/my-aqua-config/blob/main/aqua.yaml)

```yaml
registries:
- type: standard
  ref: v0.10.10 # renovate: depName=suzuki-shunsuke/aqua-registry

packages:
- name: helm/helm@v3.7.0
- name: golangci/golangci-lint@v1.42.1
```

## Configuration File Path

aqua searches the following configuration files.

1. `--config (-c)` option (environment variable `AQUA_CONFIG`
1. `\.?aqua\.ya?ml` from the current directory to the root directory. If configuration files are found in the multiple directories, aqua read all of them
1. global configuration: environment variable `AQUA_GLOBAL_CONFIG`
1. global configuration: `$AQUA_ROOT_DIR/global/\.?aqua\.ya?ml`

`aqua exec` and `aqua install -a` reads global configuration files, but otherwise aqua doesn't read global configuration files.

## Environment variables

* `AQUA_LOG_LEVEL`: (default: `info`) Log level
* `AQUA_CONFIG`: configuration file path
* `AQUA_GLOBAL_CONFIG`: global configuration file paths separated by semicolon `:`
* `AQUA_ROOT_DIR`: (default: `$HOME/.aqua`)
* `AQUA_MAX_PARALLELISM`: (default: `5`) The maximum number of packages which are installed in parallel at the same time
* `AQUA_GITHUB_TOKEN`, `GITHUB_TOKEN`: GitHub Access Token. This is required to install private repository's package

## Configuration attributes

* [registries](#registries): The list of registries
* [packages](#packages): The list of installed packages
* [inline_registry](#inline_registry): The inline registry

## `registries`

e.g.

```yaml
registries:
- type: standard
  ref: v0.10.10 # renovate: depName=suzuki-shunsuke/aqua-registry
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
  ref: v0.10.10 # renovate: depName=suzuki-shunsuke/aqua-registry
```

* `ref`: the Registry Version. Please check [Releases](https://github.com/suzuki-shunsuke/aqua-registry/releases)

This is equivalent to the following definition.

```yaml
registries:
- name: standard
  type: github_content
  repo_owner: suzuki-shunsuke
  repo_name: aqua-registry
  ref: v0.10.10 # renovate: depName=suzuki-shunsuke/aqua-registry
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
  ref: v0.10.10 # renovate: depName=suzuki-shunsuke/aqua-registry
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

* `name`: (string, optional) package name. If `import` isn't set, this is required
  * format: `<package name>[@<package version>]`
* `registry`: (string, optional) registry name
  * default value is `standard`
* `version`: (string, optional) package version
* `import`: (string, optional) glob pattern of package files. This is relative path from the configuration file. This is parsed with [filepath.Glob](https://pkg.go.dev/path/filepath#Glob). Please see [Split packages definition with `import`](#split-packages-definition-with-import) too

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
  version: v1.43.0 # renovate: depName=golangci/golangci-lint
```

If the package name in the code comment is wrong, the package version is changed wrongly.

```yaml
- name: golangci/golangci-lint
  registry: standard
  # depName is wrong!
  version: v1.42.1 # renovate: depName=helm/helm
```

On the other hand, you can prevent such a mis configuration by the first style.

### Split packages definition with `import`

e.g.

Directory structure

```
aqua.yaml
aqua/
  conftest.yaml
```

aqua.yaml

```yaml
registries:
- type: standard # standard registry
  ref: v0.10.10 # renovate: depName=suzuki-shunsuke/aqua-registry
- import: aqua/*.yaml
```

aqua/conftest.yaml

```yaml
packages:
- name: open-policy-agent/conftest@v0.28.2
```

`import` is useful for CI.
You can execute test and lint only when the specific package is updated.

ex. GitHub Actions' [`on.<push|pull_request>.paths`](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onpushpull_requestpaths)

```yaml
name: conftest
on:
  pull_request:
    paths:
    - policy/**.rego
    - aqua/conftest.yaml
```

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
