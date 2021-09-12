# aqua

[![Build Status](https://github.com/suzuki-shunsuke/aqua/workflows/test/badge.svg)](https://github.com/suzuki-shunsuke/aqua/actions)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/aqua.svg)](https://github.com/suzuki-shunsuke/aqua)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/aqua/main/LICENSE)

Declarative CLI Version manager. Support `Lazy Install` and Sharable configuration mechanism named `Registry`. Switch versions seamlessly.

## Index

* [Slide (Speaker Deck)](https://speakerdeck.com/szksh/introduction-of-aqua)
* [Tutorial](tutorial/README.md)
* [Usage](docs/usage.md)
* [Configuration](docs/config.md)
* [Registry](docs/registry.md)
* [Continuous update by Renovate](docs/renovate.md)
* [How does Lazy Install works?](docs/lazy-install.md)

## Blog

* English
  * [2021-09-08 aqua - Declarative CLI Version Manager](https://dev.to/suzukishunsuke/aqua-declarative-cli-version-manager-1ibe)
* Japanese
  * [2021-08-28 aqua - CLI ツールのバージョン管理](https://techblog.szksh.cloud/aqua/)
  * [2021-09-04 aqua v0.1.0 から v0.5.0 での変更点](https://techblog.szksh.cloud/aqua-v0.5/)
  * [2021-09-05 aqua の設定ファイルをインタラクティブに生成する generate コマンド](https://techblog.szksh.cloud/aqua-generate/)

## Note: Windows isn't supported

Currently, aqua doesn't support Windows.

## Overview

You can install CLI tools and manage their versions with declarative YAML configuration `aqua.yaml`.

e.g. Install jq, direnv, and fzf with aqua.

```yaml
registries:
- type: standard
  ref: v0.5.2 # renovate: depName=suzuki-shunsuke/aqua-registry

packages:
- name: jq
  registry: standard
  version: jq-1.6
- name: direnv
  registry: standard
  version: v2.28.0 # renovate: depName=direnv/direnv
- name: fzf
  registry: standard
  version: 0.27.2 # renovate: depName=junegunn/fzf
```

After writing the configuration, you can install them by `aqua i`.

```
$ aqua i
```

`aqua i` installs all packages all at once.
Tools are installed in `~/.aqua/pkgs` and symbolic links are created in `~/.aqua/bin`, so please add `~/.aqua/bin` to the environment variable `PATH`.

It takes a long time to install many tools all at once, and some tools might not be actually needed.

So instead of `aqua i` let's execute `aqua i -l`.

```
$ aqua i -l
```

`aqua i -l` creates symbolic links to aqua-proxy in `~/.aqua/bin` but skipping the downloading and installing tools.
When you execute the tool, the tool is installed automatically if it isn't installed yet before it is executed.
We call this feature as _lazy install_.
By the lazy install, you don't have to execute aqua explicitly after changing the tool's version.
When `aqua.yaml` is managed with Git, the lazy install is very useful because `aqua.yaml` is updated by `git pull` then the update is reflected automatically.

By adding `aqua.yaml` in your Git repositories, you can manage tools per repository.
You can change the version of tools per project.

aqua installs the tools in the shared directory `~/.aqua`,
so the same version of the same tool is installed only at once.
It saves the time and the disk usage.

aqua supports the mechanism named `Registry`.
You can share and reuse the aqua configuration, so it makes easy to write `aqua.yaml`.

```yaml
registries:
- type: standard
  ref: v0.5.2 # renovate: depName=suzuki-shunsuke/aqua-registry

packages:
- name: direnv
  registry: standard
  version: v2.28.0 # renovate: depName=direnv/direnv
```

In the above configuration, [the standard Registry](https://github.com/suzuki-shunsuke/aqua-registry/blob/main/registry.yaml) is used so you can install direnv easily.

By the command `aqua generate`, you can check if the registry supports the tool you need and write the configuration quickly.

```
$ aqua g
```

`aqua g` launches the interactive UI and you can select the package and it's version interactively.

```
  direnv (standard)
  consul (standard)
  conftest (standard)
> golangci-lint (standard)
  47/47
>
```

If the selected package type is `github_release`, you can select the package version interactively.

```
  v1.40.0
  v1.40.1
  v1.41.0
  v1.41.1
> v1.42.0
  30/30
>
```

After selecting the package and its version, the configuration is outputted.

```console
$ aqua g
- name: golangci-lint
  registry: standard
  version: v1.42.0
```

If the Registries don't support the tool, you can send the pull request to the registry or create your own Registry or add the configuration in `aqua.yaml` as `inline` Registry.

## Quick Start

Install aqua.

```console
$ curl -sSfL https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/v0.1.3/aqua-installer | bash -s -- -i bin/aqua
$ export PATH=$PWD/bin:$HOME/.aqua/bin:$PATH
$ export GITHUB_TOKEN=<your personal access token>
```

Write `aqua.yaml`.

```yaml
packages:
- name: jq
  registry: standard
  version: jq-1.5

registries:
- type: standard
  ref: v0.5.2 # renovate: depName=suzuki-shunsuke/aqua-registry
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
* Ecosystem named `Registry` - it eases to write aqua configuration
  * You can share and reuse the aqua configuration. We provide the standard registry too

## Install

Please download a binary from the [Release Page](https://github.com/suzuki-shunsuke/aqua/releases).

Or you can install aqua quickly with [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer).

e.g.

```
$ curl -sSfL https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/v0.1.3/aqua-installer | bash
```

GitHub Actions

e.g.

```yaml
- uses: suzuki-shunsuke/aqua-installer@v0.1.3
  with:
    version: v0.3.1
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
  registries/
    github_content/
      github.com/
        suzuki-shunsuke/
          aqua-registry/
            v0.1.1-0/
              registry.yaml
```

## Related Projects

* [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy)
* [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer): Install aqua quickly
* [aqua-registry](https://github.com/suzuki-shunsuke/aqua-registry): Standard Registry
* [aqua-renovate-config](https://github.com/suzuki-shunsuke/aqua-renovate-config): Renovate Configuration to update packages and registries

## Example

* [suzuki-shunsuke/my-aqua-config](https://github.com/suzuki-shunsuke/my-aqua-config)
* [suzuki-shunsuke/example-aqua](https://github.com/suzuki-shunsuke/example-aqua)

## Change Log

Please see [Releases](https://github.com/suzuki-shunsuke/aqua/releases).

## Versioning Policy

We are Conforming [suzuki-shunsuke/versioning-policy v0.1.0](https://github.com/suzuki-shunsuke/versioning-policy/blob/v0.1.0/POLICY.md), which is compatible with [Semantic Versioning 2.0.0](https://semver.org/).

## License

[MIT](LICENSE)
