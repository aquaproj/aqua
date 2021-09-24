# aqua

[![Build Status](https://github.com/suzuki-shunsuke/aqua/workflows/test/badge.svg)](https://github.com/suzuki-shunsuke/aqua/actions)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/aqua.svg)](https://github.com/suzuki-shunsuke/aqua)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/aqua/main/LICENSE)

Declarative CLI Version manager. Support `Lazy Install` and Sharable configuration mechanism named `Registry`. Switch versions seamlessly.

[Config](docs/config.md) | [Usage](docs/usage.md) | [Document](docs) | [Release Note](https://github.com/suzuki-shunsuke/aqua/releases)

**Currently, aqua doesn't support Windows.**

## Feature

* Install tools easily by the declarative YAML Configuration and simple command `aqua i`
* Unify tool versions at local development environment and CI/CD for the project
  * It solves the problem due to the difference of versions
* Support using defferent versions per project
* Lazy Install
* Registry - Sharable Configuration mechanism
  * Welcome your contribution to [Standard Registry](https://github.com/suzuki-shunsuke/aqua-registry)!
* Save installation time and disk by sharing tools across projects
* Set up your laptops quickly

## Overview

You can install CLI tools and manage their versions with declarative YAML configuration `aqua.yaml`.

e.g. Install jq, direnv, and fzf with aqua.

```yaml
registries:
- type: standard
  ref: v0.8.6 # renovate: depName=suzuki-shunsuke/aqua-registry

packages:
- name: stedolan/jq
  registry: standard
  version: jq-1.6
- name: direnv/direnv
  registry: standard
  version: v2.28.0 # renovate: depName=direnv/direnv
- name: junegunn/fzf
  registry: standard
  version: 0.27.2 # renovate: depName=junegunn/fzf
```

You can install tools by `aqua i`.

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
You can use the different version of tools per project.

aqua installs the tools in the shared directory `~/.aqua`,
so the same version of the same tool is installed only at once.
It saves the installation time and the disk usage.

aqua supports the mechanism named `Registry`.
You can share and reuse the aqua configuration, so it makes easy to write `aqua.yaml`.

```yaml
registries:
- type: standard
  ref: v0.8.6 # renovate: depName=suzuki-shunsuke/aqua-registry

packages:
- name: direnv/direnv
  registry: standard
  version: v2.28.0 # renovate: depName=direnv/direnv
```

In the above configuration, [the Standard Registry](https://github.com/suzuki-shunsuke/aqua-registry/blob/main/registry.yaml) is used so you can install direnv easily.

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

After selecting the package, the configuration is outputted.

```console
$ aqua g
- name: golangci-lint
  registry: standard
  version: v1.42.0
```

If the Registries don't support the tool, you can send the pull request to the registry or create your own Registry or add the configuration in `aqua.yaml` as `inline` Registry.

## Install

Please download a binary from the [Release Page](https://github.com/suzuki-shunsuke/aqua/releases). Or you can install aqua quickly with [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer).

e.g.

```console
$ curl -sSfL \
  https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/v0.1.3/aqua-installer |
  bash -s -- -i ~/bin/aqua -v v0.7.1 # renovate: depName=suzuki-shunsuke/aqua
```

GitHub Actions

e.g.

```yaml
- uses: suzuki-shunsuke/aqua-installer@v0.1.3
  with:
  version: v0.7.1 # renovate: depName=suzuki-shunsuke/aqua
    install_path: /tmp/bin/aqua
```

[int128/aqua-action](https://github.com/int128/aqua-action)

```yaml
- uses: int128/aqua-action@v1
  with:
    version: v0.7.1 # renovate: depName=suzuki-shunsuke/aqua
```

aqua requires the environment variable `GITHUB_TOKEN`, which is GitHub Access Token.
Add `~/.aqua/bin` to the environmenet variable `PATH`.

```console
$ export GITHUB_TOKEN=xxx
$ export PATH=$HOME/.aqua/bin:$PATH
```

## Install tools in your laptop

If you want to install tools in your laptop regardless specific project,
create the global configuration `~/.aqua/global/aqua.yaml`.
Like dotfiles, it is good to manage the Global Configuration with Git and share it with your multiple laptops.

For example,

```
$ git clone https://github.com/suzuki-shunsuke/my-aqua-config ~/.aqua/global
$ cd ~/.aqua/global
$ aqua i -l
```

## Related Projects

* [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy)
* [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer): Install aqua quickly
* [aqua-registry](https://github.com/suzuki-shunsuke/aqua-registry): Standard Registry
* [aqua-renovate-config](https://github.com/suzuki-shunsuke/aqua-renovate-config): Renovate Configuration to update packages and registries
* Third Party Projects
  * [int128/aqua-action](https://github.com/int128/aqua-action) - Action to install packages using aqua

## Welcome your contribution to Standard Registry!

https://github.com/suzuki-shunsuke/aqua-registry

If tools you want aren't found in [Standard Registry](https://github.com/suzuki-shunsuke/aqua-registry), please create issues or send pull requests!

## Slide, Blog

* English
  * [2021-09-08 aqua - Declarative CLI Version Manager](https://dev.to/suzukishunsuke/aqua-declarative-cli-version-manager-1ibe)
  * [2021-09-02 Slide - Introduction of aqua](https://speakerdeck.com/szksh/introduction-of-aqua)
* Japanese
  * [2021-08-28 aqua - CLI ツールのバージョン管理](https://techblog.szksh.cloud/aqua/)
  * [2021-09-04 aqua v0.1.0 から v0.5.0 での変更点](https://techblog.szksh.cloud/aqua-v0.5/)
  * [2021-09-05 aqua の設定ファイルをインタラクティブに生成する generate コマンド](https://techblog.szksh.cloud/aqua-generate/)

## Versioning Policy

We are Conforming [suzuki-shunsuke/versioning-policy v0.1.0](https://github.com/suzuki-shunsuke/versioning-policy/blob/v0.1.0/POLICY.md), which is compatible with [Semantic Versioning 2.0.0](https://semver.org/).

## License

[MIT](LICENSE)
