# Configuration (registry.yaml)

e.g. [registry.yaml](https://github.com/suzuki-shunsuke/aqua-registry/blob/main/registry.yaml)

```yaml
packages:
# init: a
- type: github_release
  repo_owner: accurics
  repo_name: terrascan
  asset: 'terrascan_{{trimV .Version}}_{{title .OS}}_{{.Arch}}.tar.gz'
  link: https://docs.accurics.com/projects/accurics-terrascan/en/latest/
  description: Detect compliance and security violations across Infrastructure as Code to mitigate risk before provisioning cloud native infrastructure
  replacements:
    amd64: x86_64
```

* `packages`: The list of packages

## Package types

* [github_release](#github_release-package): The package is downloaded from GitHub Releases
* [http](#http-package): The package is donwloaded from the specified URL

## Package's Common attributes

* `type`: (string, required) the package type
  * `github_release`
  * `http`
* `name`: (string) the package name. This is used to specify the package in `aqua.yaml`. name must be unique in the same registry
* [files](#files): The list of executable files
* `format`: The asset format (e.g. `zip`, `tar.gz`). This is used to unarchive or decompress the asset. If this isn't specified, aqua tries to specify the format from the file extenstion. If the file isn't archived and isn't compressed, please specify `raw`
* `link`: URL about the package. This is used for `aqua g`
* `description`: The description about the package. This is used for `aqua g`
* [replacements](#replacements-format_overrides): A map which is used to replace some Template Variables like `OS` and `Arch`
* [format_overrides](#replacements-format_overrides): A list of the pair OS and the asset format

### `files`

* `name`: the file name
* `src`: (default: the value of `name`, type: `template string`) the path to the file from the archive file's root.

## `github_release` Package

* `repo_owner`: The repository owner name
* `repo_name`: The repository name
* `asset`: The template string of GitHub Release's asset name
  * e.g. `'lima-{{trimV .Version}}-{{title .OS}}-{{.Arch}}.tar.gz'`

## `http` Package

* `url`: The template string of URL where the package is downloaded
  * e.g. `'https://storage.googleapis.com/kubernetes-release/release/{{.Version}}/bin/{{.OS}}/{{.Arch}}/kubectl'`

## `replacements`, `format_overrides`

These attributes are inspired to [goreleaser's Archive](https://goreleaser.com/customization/archive/).
If the package is released with [goreleaser](https://goreleaser.com/),
you may copy and paste `replacements` and `format_overrides` from `.goreleaser.yaml`.

e.g.

* replacements
  * [goreleaser.yml](https://github.com/aquasecurity/trivy/blob/v0.19.2/goreleaser.yml#L62-L73)
  * [registry.yaml](https://github.com/suzuki-shunsuke/aqua-registry/blob/v0.8.0/registry.yaml#L44-L55)
* format_overrides
  * [.goreleaser.yml](https://github.com/iawia002/annie/blob/v0.11.0/.goreleaser.yml#L51-L54)
  * [registry.yaml](https://github.com/suzuki-shunsuke/aqua-registry/blob/v0.8.0/registry.yaml#L361-L364)

## Default values of `github_release` package

* `name`: `<repo owner>/<repo name>`
* `link`: `https://github.com/<repo owner>/<repo name>`
* `files`: `[{"name":"<repo name>"}]`

For example, in case of `weaveworks/eksctl` the following default values are set.

```yaml
name: weaveworks/eksctl
link: https://github.com/weaveworks/eksctl
files:
- name: eksctl
```

## Template String

Some fields are parsed with [Go's text/template](https://pkg.go.dev/text/template) and [sprig](http://masterminds.github.io/sprig/).

### Common Template Functions

* `trimV`: This is equivalent to `trimPrefix "v"`

### Template Variables

* `OS`: A string which `GOOS` is replaced by `replacements`. If `replacements` isn't set, `OS` is equal to `GOOS`. Basically you should use `OS` for the consistency
* `Arch`: A string which `GOARCH` is replaced by `replacements`. If `replacements` isn't set, `Arch` is equal to `GOARCH`. Basically you should use `OS` for the consistency
* `GOOS`: Go's [runtime.GOOS](https://pkg.go.dev/runtime#pkg-constants)
* `GOARCH`: Go's [runtime.GOARCH](https://pkg.go.dev/runtime#pkg-constants)
* `Version`: Package `version`
* `Format`: Package `format`
* `FileName`: `files[].name`
