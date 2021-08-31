# aqua

[![Build Status](https://github.com/suzuki-shunsuke/aqua/workflows/test/badge.svg)](https://github.com/suzuki-shunsuke/aqua/actions)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/aqua.svg)](https://github.com/suzuki-shunsuke/aqua)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/aqua/main/LICENSE)

Version manager of CLI.

## Quick Start

Install aqua.

```console
$ curl -sSfL https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/v0.1.2/aqua-installer | bash -s -- -i bin/aqua
$ export PATH=$PWD/bin:$HOME/.aqua/bin:$PATH
$ export GITHUB_TOKEN=<your personal access token>
```

Write `aqua.yaml`.

```yaml
packages:
- name: jq
  registry: inline
  version: jq-1.5
inline_registry:
- name: jq
  type: github_release
  repo_owner: stedolan
  repo_name: jq
  asset: 'jq-{{if eq .OS "darwin"}}osx{{else}}{{.OS}}{{end}}-{{.Arch}}'
  files:
  - name: jq
```

Install tools.

```console
$ aqua i
```

Tools are installed successfully.

```console
$ jq --version
jq-1.5
```

Edit `aqua.yaml`.

```
$ sed -i "s/jq-1\.5/jq-1.6/" aqua.yaml
```

Run `jq` again, then jq's new version is installed automatically and `jq` is run.

```
$ jq --version
jq-1.6
```

## Index

* [Tutorial](tutorial/README.md)
* [Usage](docs/usage.md)
* [Configuration](docs/config.md)

## Main Usecase

* Install tools in CI/CD
  * [Example](https://github.com/suzuki-shunsuke/example-aqua#install-tools-in-cicd)
* Install tools for your project's local development
  * [Example](https://github.com/suzuki-shunsuke/example-aqua#install-tools-for-this-projects-local-development)
* Install tools in your laptop
  * [Example](https://github.com/suzuki-shunsuke/my-aqua-config)

## Feature

* Declarative YAML Configuration
  * You don't have to execute commands imperatively to install tools
* Manage versions per project
  * You can change tools version per project
* Install tools when they are executed
  * When you execute the tool which isn't installed yet, aqua installs the tool and execute the tool
* Share tools across projects
  * aqua installs tools in the shared directory `~/.aqua`. It saves time and disk to install tools

## Install

Please download a binary from the [Release Page](https://github.com/suzuki-shunsuke/aqua/releases).

Or you can install aqua quickly with [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer).

e.g.

```
$ curl -sSfL https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/v0.1.2/aqua-installer | bash
```

GitHub Actions

e.g.

```yaml
- uses: suzuki-shunsuke/aqua-installer@v0.1.2
  with:
    version: v0.1.0-9
    install_path: /tmp/bin/aqua
```

## Where are tools installed?

* Symbolic links are created in `$HOME/.aqua/bin`, so add this to the environment variable `PATH`
* Tools are installed in `$HOME/.aqua/pkgs`

```
(your working directory)/
  aqua.yaml
~/.aqua/ # $AQUA_ROOT_DIR (default ~/.aqua)
  bin/
    aqua-proxy (symbolic link to aqua-proxy)
    <tool> (symbolic link to aqua-proxy)
  global/
    aqua.yaml # global configuration
  pkgs/
    github_release/
      github.com/
        suzuki-shunsuke/
          aqua-proxy/
            v0.1.0/
              aqua-proxy_darwin_amd64.tar.gz
                aqua-proxy
```

## Related Projects

* [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy)
* [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer)

## Example

* [suzuki-shunsuke/my-aqua-config](https://github.com/suzuki-shunsuke/my-aqua-config)
* [suzuki-shunsuke/example-aqua](https://github.com/suzuki-shunsuke/example-aqua)

## Change Log

Please see [Releases](https://github.com/suzuki-shunsuke/aqua/releases).

## Blog

* [aqua - CLI ツールのバージョン管理](https://techblog.szksh.cloud/aqua/)

## License

[MIT](LICENSE)
