# Configuration

e.g. [tutorial/aqua.yaml](../tutorial/aqua.yaml), [suzuki-shunsuke/my-aqua-config](https://github.com/suzuki-shunsuke/my-aqua-config/blob/main/aqua.yaml), [suzuki-shunsuke/aqua-registry](https://github.com/suzuki-shunsuke/aqua-registry/blob/main/registry.yaml)

## Configuration File Path

The configuration file path can be specified with the `--config (-c)` option.
If the confgiuration file path isn't specified, the file named `[.]aqua.y[a]ml`  would be searched from the current directory to the root directory.

## Environment variables

* `AQUA_LOG_LEVEL`: (default: `info`) Log level
* `AQUA_CONFIG`: configuration file path
* `AQUA_ROOT_DIR`: (default: `$HOME/.aqua`)
* `AQUA_MAX_PARALLELISM`: (default: `5`) The maximum number of packages which are installed in parallel at the same time
* `GITHUB_TOKEN`: GitHub Access Token. This is required to install `github_release` packages

## Configuration File Format

* `packages`: The list of installed packages
* `inline_registry`: The inline registry
* `registries`: The list of registries

### type: Inline Registry

* `packages`: The list of package metadata

### type: Package

* `name`: the package name. This is used to map the package and the package metadata
  `name` must be unique in the same [Registry](#registry)
* `registry`: the name of package metadata
* `version`: the package version

### type: PackageInfo

PackageInfo is the package metadata how the package is installed.

* `name`: the package name
* `type`: the package type. Either `github_release` or `http` is supported
* `format`: the archive type (e.g. `zip`, `tar.gz`). Basically you don't have to specify this field because `aqua` understand the archive type from the filename extension.
  If the `format` is `raw` or the filename has no extension, `aqua` treats the file isn't archived and uncompressed.
* `format_overrides`
* `replacements`
* `description`
* `link`
* `files`: The list of files which the archive includes

The type `github_release` has the following fields.

* `repo_owner`: GitHub Repository owner
* `repo_name`: GitHub Repository name
* `asset`: (type: `template string`) GitHub Release asset name

The type `http` has the following fields.

* `url`: Downloaded URL

### Registry

`Registry` is the list of package metadata.

* `name`: registy name
* `type`: registy type. `github_content` and `local` and `standard` are supported

`github_content` registry has the following fields.

* `repo_owner`: GitHub Repository owner
* `repo_name`: GitHub Repository name
* `ref`: GitHub Content ref (e.g. `v0.1.0`)
* `path`: GitHub Content file path. This must be a file

`local` registry has the following fields.

* `path`: The registry file path. This is either the absolute path or the relative path from aqua configuration file

`standard` registry is special registry.

Only `ref` field is supported.

```yaml
registries:
- type: standard
  ref: v0.7.0 # renovate: depName=suzuki-shunsuke/aqua-registry
```

This is equivalent to the following definition.

```yaml
registries:
- name: standard
  type: github_content
  repo_owner: suzuki-shunsuke
  repo_name: aqua-registry
  ref: v0.7.0 # renovate: depName=suzuki-shunsuke/aqua-registry
  path: registry.yaml
```

### type: File

* `name`: the file name
* `src`: (default: the value of `name`, type: `template string`) the path to the file from the archive file's root.

### template string

Some fields are parsed with [Go's text/template](https://pkg.go.dev/text/template) and [sprig](http://masterminds.github.io/sprig/).

* `PackageInfo.asset`
* `PackageInfo.url`
* `File.src`

#### `PackageInfo.url`, `PackageInfo.asset`

The following variables are passed to the template.

* `OS`
* `Arch`
* `GOOS`: Go's [runtime.GOOS](https://pkg.go.dev/runtime#pkg-constants)
* `GOARCH`: Go's [runtime.GOARCH](https://pkg.go.dev/runtime#pkg-constants)
* `Version`
* `Format`

#### `File.src`

The following variables are passed to the template.

* `OS`
* `Arch`
* `GOOS`: Go's [runtime.GOOS](https://pkg.go.dev/runtime#pkg-constants)
* `GOARCH`: Go's [runtime.GOARCH](https://pkg.go.dev/runtime#pkg-constants)
* `Version`
* `Format`
* `FileName`
