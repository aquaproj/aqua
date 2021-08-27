# Configuration

e.g. [tutorial/aqua.yaml](../tutorial/aqua.yaml)

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
* `inline_repository`: The list of package metadata

### type: Package

* `name`: the package name. This is used to map the package and the package metadata
  `name` must be unique in the same [Repository](#repository)
* `repository`: the name of package metadata
* `version`: the package version

### type: PackageInfo

PackageInfo is the package metadata how the package is installed.

* `name`: the package name
* `type`: the package type. Only `github_release` is supported
* `archive_type`: the archive type (e.g. `zip`, `tar.gz`). Basically you don't have to specify this field because `aqua` understand the archive type from the filename extension.
  If the `archive_type` is `raw` or the filename has no extension, `aqua` treats the file isn't archived and uncompressed.
* `files`: The list of files which the archive includes

`github_release` has the following fields.

* `repo_owner`: GitHub Repository owner
* `repo_name`: GitHub Repository name
* `artifact`: (type: `template string`) GitHub Release asset name

### Repository

`Repository` is the list of package metadata.
Only `inline` repository is supported.

### type: File

* `name`: the file name
* `src`: (default: the value of `name`, type: `template string`) the path to the file from the archive file's root.

### template string

Some fields are parsed with [Go's text/template](https://pkg.go.dev/text/template) and [sprig](http://masterminds.github.io/sprig/).

* `PackageInfo.artifact`
* `File.src`
