---
sidebar_position: 200
---

# Configuration

e.g.

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry

packages:
- name: helm/helm@v3.7.0
- name: golangci/golangci-lint@v1.42.1
- import: aqua/*.yaml
```

## Configuration File Path

aqua searches the following configuration files.

1. `--config (-c)` option (environment variable `AQUA_CONFIG`)
1. `\.?aqua\.ya?ml` or `\.?aqua/aqua\.ya?ml` from the current directory to the root directory. If configuration files are found in the multiple directories, aqua read all of them
1. global configuration: environment variable `AQUA_GLOBAL_CONFIG`

To install tools in global configuration files, you have to set `-a` to `aqua install` command.

:::info
[From aqua v1.33.0, aqua supports keeping aqua's configuration files in one directory `.?aqua`](/docs/guides/keep-in-one-dir).
:::

## Environment variables

* `AQUA_LOG_LEVEL`: (default: `info`) Log level
* `AQUA_CONFIG`: configuration file path
* [AQUA_GLOBAL_CONFIG](/docs/tutorial/global-config): global configuration file paths separated by semicolon `:`
* `AQUA_POLICY_CONFIG`: [policy file](/docs/reference/security/policy-as-code) paths separated by semicolon `:`
* [`AQUA_DISABLE_COSIGN`: `aqua >= v2.22.0` If true, the verification with Cosign is disabled](/docs/reference/security/cosign-slsa#disable-cosign-and-slsa-aqua-installer)
* [`AQUA_DISABLE_SLSA`: `aqua >= v2.22.0` If true, the verification with SLSA Provenance is disabled](/docs/reference/security/cosign-slsa#disable-cosign-and-slsa-aqua-installer)
* [`AQUA_DISABLE_GITHUB_ARTIFACT_ATTESTATION`: `aqua >= v2.35.0` If true, the verification using GitHub Artifact Attestations is disabled](/docs/reference/security/github-artifact-attestations#disable-the-verification-of-github-artifact-attestations)
* `AQUA_DISABLE_POLICY`: If true, [Policy](/docs/reference/security/policy-as-code) is disabled (aqua >= v2.1.0)
* `AQUA_DISABLE_LAZY_INSTALL`: If true, [Lazy Install](/docs/reference/lazy-install/) is disabled (aqua >= v2.9.0)
* `AQUA_ROOT_DIR`: The directory path where aqua install tools
  * default (linux and macOS): `${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua`
  * default (windows): `${HOME/AppData/Local}/aquaproj-aqua`
* `AQUA_MAX_PARALLELISM`: (default: `5`) The maximum number of packages which are installed in parallel at the same time
* `AQUA_GITHUB_TOKEN`, `GITHUB_TOKEN`: GitHub Access Token. This is required to install private repository's package
  * [You can also manage GitHub access tokens using ghtkn integration](/docs/reference/security/ghtkn)
  * [You can also manage a GitHub access token using Keyring](/docs/reference/security/keyring)
* [AQUA_GHTKN_ENABLED](/docs/reference/security/ghtkn) `aqua >= v2.54.0`
* [AQUA_KEYRING_ENABLED](/docs/reference/security/keyring) `aqua >= v2.51.0`
* [AQUA_LOG_COLOR](log-color.md): Log color setting (`always|auto|never`)
* [AQUA_PROGRESS_BAR](progress-bar.md): The progress bar is disabled by default, but you can enable it by setting the environment variable `AQUA_PROGRESS_BAR` to `true`
* [AQUA_GOOS, AQUA_GOARCH](/docs/develop-registry/change-os-arch-for-test)
* [AQUA_X_SYS_EXEC](/docs/reference/execve-2)
* (Deprecated) [AQUA_EXPERIMENTAL_X_SYS_EXEC](experimental-feature.md#aqua_experimental_x_sys_exec)
* `AQUA_GENERATE_WITH_DETAIL`: (boolean, default: `false`) If true, aqua outputs additional information such as description and link [#2027](https://github.com/orgs/aquaproj/discussions/2027) [#2062](https://github.com/aquaproj/aqua/pull/2062) (aqua >= v2.9.0)

```console
$ env AQUA_GENERATE_WITH_DETAIL=true aqua g cli/cli
- name: cli/cli@v2.2.0
  description: GitHub’s official command line tool
  link: https://github.com/cli/cli
```

* `AQUA_REMOVE_MODE`: [`aqua remove` command's `-mode` option](/docs/guides/uninstall-packages)

## JSON Schema

* https://github.com/aquaproj/aqua/tree/main/json-schema
* https://github.com/aquaproj/aqua/blob/main/json-schema/aqua-yaml.json
* https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json

### Input Complementation by YAML Language Server

Add a code comment to aqua.yaml:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json
```

If you specify a branch like `main` as version, editors can't reflect the update of JSON Schema well as they cache JSON Schema.
You would need to do something like reopening the file.
So it's good to specify semver and update it periodically.

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/v2.40.0/json-schema/aqua-yaml.json
```

Using Renovate and [our Renovate Config Preset](https://github.com/suzuki-shunsuke/renovate-config/blob/main/yaml-language-server.json), you can automate the update:

```json
{
  "extends": [
    "github>suzuki-shunsuke/renovate-config:yaml-language-server#3.0.0"
  ]
}
```

e.g. https://github.com/aquaproj/aqua-installer/pull/735

## Configuration attributes

* [registries](#registries): The list of registries
* [packages](#packages): The list of installed packages
* [checksum](checksum.md): configuration for checksum verification
* [import_dir](/docs/guides/split-config#import_dir): A directory path where files are imported. `aqua >= v2.44.0`

## `registries`

e.g.

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry
```

Registry types

* [standard](#standard-registry): aqua's [Standard Registry](https://github.com/aquaproj/aqua-registry)
* [local](#local-registry): local file
* [github_content](#github_content-registry): Get the registry by GitHub Repository Content API
* [http](#http-registry): Get the registry from HTTP(S) endpoints (aqua >= v2.56.0)

### `standard` registry

e.g.

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry
```

* `ref`: the Registry Version. Please check [Releases](https://github.com/aquaproj/aqua-registry/releases)

This is equivalent to the following definition.

```yaml
registries:
- name: standard
  type: github_content
  repo_owner: aquaproj
  repo_name: aqua-registry
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry
  path: registry.yaml
```

You can also specify a commit hash as `ref`.

```yaml
registries:
- type: standard
  ref: 0d1572334a460e5a74f2a6455e510d8a4d6c8e93
```

:::caution
Don't specify a branch name as `ref`, because aqua treats the ref as immutable.

```yaml
registries:
- type: standard
  ref: main # Specify a tag or commit hash
```

:::

### `local` registry

e.g.

```yaml
registries:
- name: local
  type: local
  path: registry.yaml
- name: home
  type: local
  path: $HOME/aqua-registry.yaml
```

* `name`: Registry name
* `path`: The file path. Either absolute path or relative path from `aqua.yaml`. If `path` starts with `$HOME` + `OS specific path separator such as '/'`, it's replaced with the home directory path

Please see [Configuration (registry.yaml)](/docs/reference/registry-config).

### `github_content` registry

e.g.

```yaml
registries:
- name: foo
  type: github_content
  repo_owner: aquaproj
  repo_name: aqua-registry
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry
  path: registry.yaml
```

* `name`: Registry Name
* `repo_owner`: Repository Owner name
* `repo_name`: Repository name
* `ref`: Repository tag or commit hash. Don't specify a branch name as `ref`, because aqua treats the ref as immutable
* `path`: file path from the repository root directory

### `http` registry

:::info
aqua >= v2.56.0
:::

The `http` registry type allows you to load registries from HTTP(S) endpoints, enabling private registries hosted on internal HTTP servers.

e.g.

```yaml
registries:
  - name: my-registry
    type: http
    url: https://my.registry.com/team/{{.Version}}/registry.yaml
    version: v1.0.0
    format: raw
```

Required attributes:

* `name`: Registry name
* `url`: HTTP(S) URL to fetch the registry. Must contain `{{.Version}}` template placeholder for determinism
* `version`: Registry version to use

Optional attributes:

* `format`: Registry file format. Either `raw` (default, for YAML/JSON) or `tar.gz` (for archived registries)

The `url` supports [Go templates](/docs/reference/registry-config/template) with the `{{.Version}}` placeholder and [Sprig functions](http://masterminds.github.io/sprig/):

```yaml
registries:
  - name: my-registry
    type: http
    url: https://my.registry.com/registries/{{trimV .Version}}/registry.yaml
    version: v1.0.0
```

This renders as: `https://my.registry.com/registries/1.0.0/registry.yaml`

**Security**: The version field is validated to prevent path traversal attacks (e.g., `../` is not allowed).

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
* `go_version_file`: (string, optional) `aqua >= v2.28.0` [#2632](https://github.com/aquaproj/aqua/pull/2632) A file path to go.mod or go.work. This field is used to get the version of go from go directive in go.mod or go.work
* `version_expr`: (string, optional) `aqua >= v2.40.0` [Please see here for details](version-expr.md)
* `version_expr_prefix`: (string, optional) `aqua >= v2.40.0` [Please see here for details](version-expr.md)
* `import`: (string, optional) glob pattern of package files. This is relative path from the configuration file. This is parsed with [filepath.Glob](https://pkg.go.dev/path/filepath#Glob). Please see [Split the list of packages](/docs/guides/split-config) too.
* [tags](/docs/guides/package-tag): Filter installed packages. Please see [Filter packages with tags](/docs/guides/package-tag)
* `update`: The setting for `aqua update` command
  * `update.enabled`: If this is false, `aqua update` command ignores the package. If the package name is passed to aqua up command explicitly, enabled is ignored. By default, enabled is true.
* `vars`: (map of string) [v2.31.0](https://github.com/aquaproj/aqua/releases/tag/v2.31.0) [#3052](https://github.com/aquaproj/aqua/pull/3052). Please see [here](/docs/reference/registry-config/vars)
* `command_aliases`: (array of objects, optional) [v2.37.0](https://github.com/aquaproj/aqua/releases/tag/v2.37.0) [#3224](https://github.com/aquaproj/aqua/pull/3224): Aliases of commands. Please see [here](/docs/guides/command-alias)

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

On the other hand, you can prevent such a miss configuration by the first style.
