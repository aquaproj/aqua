---
sidebar_position: 100
---

# Contributing

How to contribute to Standard Registry. https://github.com/aquaproj/aqua-registry

## See also

- :star: [OSS Contribution Guide](https://github.com/suzuki-shunsuke/oss-contribution-guide)
- :star: [Registry Style Guide](/docs/develop-registry/registry-style-guide)
- [Registry Configuration](/docs/reference/registry-config/)
- [Change `GOOS` and `GOARCH` for testing](/docs/develop-registry/change-os-arch-for-test)

## Changelog of document and development workflow

[change log](changelog.md).

- [2025-01-19 Add `Note of Programing Language Support`](#note-of-programing-language-support)
- [2024-12-14 Remove `cmdx new` from the guide](changelog.md#2024-12-14)
- [2024-05-24 The behaviour of `cmdx s`, `cmdx t`, and `cmdx new` were changed.](changelog.md#2024-05-24)

## Should you create an Issue before sending a Pull Request?

Basically, you don't have to create an Issue before sending a Pull Request.
But if the pull request requires the discussion before reviewing, you have to create an Issue in advance.

For example, you don't have to create an Issue in the following cases.

- Add a package
- Fix a typo

On the other hand, for example if you want to change the directory structure in `pkgs` or the workflow adding a package,
you have to create an Issue and describe what is changed and why the change is needed.

## aqua can't support some tools' plugin mechanism

Some tools have the plugin mechanism.

e.g.

- [GitHub CLI Extension](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions)
- [Terraform provider](https://developer.hashicorp.com/terraform/language/providers)
- [Gauge plugin](https://docs.gauge.org/plugin.html?os=macos&language=java&ide=null)
- etc

aqua simply installs commands in PATH (`AQUA_ROOT_DIR/bin`), but some of these plugins expect to be installed in the other location.
If aqua can't support the plugin, we will reject the pull request adding the plugin to aqua-registry.

So if you send a pull request adding a plugin to aqua-registry, please check if aqua can support the plugin.
We aren't necessarily familiar with the plugin, so please explain where the plugin expects to be installed and how the plugin works in the pull request description.

If you don't know well, please create a pull request and consult us.

## Note of Programing Language Support

aqua supports several programing languages such as Go and Node.js, but when we support a programing language, we need to be careful about where the programing language installs libraries and commands.

For instance, if the programing language installs commands in the same directory with the programing language itself, aqua can't add them to $PATH, meaning we can't execute them.
aqua doesn't support changing $PATH dynamically (We have no plan to support it as it makes aqua more complicated).
Node.js's `npm i -g` installs the same directory with node by default, so we gave up the support of Node.js before (Now aqua supports Node.js again because we can change the install path by `NPM_CONFIG_PREFIX`).
If the language installs libraries in the same directory with it, the language can't refer installed libraries when we change the version of the language.

So before supporting a programing language, we should consider carefully if it really works well.
Many programing languages have dedicated version managers, so maybe they are more appropriate.

## Requirements

- [aqua](https://aquaproj.github.io/docs/install)
- Docker

Please use the latest version.

### Commit Signing

All commits of pull requests must be signed.
Please see [the document](https://github.com/suzuki-shunsuke/oss-contribution-guide/blob/main/docs/commit-signing.md).

## Set up

1. [Please fork aquaproj/aqua-registry](https://github.com/aquaproj/aqua-registry/fork).
1. Checkout the repository
2. Run `aqua i -l` in the repository root directory to install tools which are required for contribution.

```sh
aqua i -l
```

## cmdx - Task Runner

We use [cmdx](https://github.com/suzuki-shunsuke/cmdx) as a task runner.
cmdx is installed by [Set up](#set-up) already.
We also use Docker to run tests in a container.
Please run `cmdx help` and `cmdx help <task>` to show the help.

```sh
cmdx help
cmdx help scaffold
```

### cmdx s - Scaffold configuration and test it in containers

`cmdx s <package name>` generates a configuration file `pkgs/<package name>/registry.yaml` and a test data file `pkgs/<package name>/pkg.yaml`, and tests them in containers.
It gets data from GitHub Releases by GitHub API.
By default, it gets all releases, so it takes a bit long time if the repository has a lot of releases.
[`cmdx s` isn't perfect, but you must use it when you add new packages.](#use-cmdx-s-definitely)

### cmdx t - Test a package in containers

`cmdx t [<package name>]` tests a package in containers.
If the branch name is `feat/<package name>`, you can omit the argument `<package name>`.
`cmdx t` copies files `pkgs/<package name>/{pkg.yaml,registry.yaml}` in containers and test them.
If the test succeeds, `registry.yaml` is updated.

### cmdx rm - Remove containers

`cmdx rm` removes containers.
`cmdx s` and `cmdx t` reuse containers, but if you want to test packages in clean environment, you can do it by removing containers.

### cmdx rmp - Remove an installed package from containers

`cmdx rmp [<package name>]` removes an installed package from containers.
If the branch name is `feat/<package name>`, you can omit the argument `<package name>`.
It runs `aqua rm <package name>` and removes `aqua-checksums.json` in containers.
This task is useful when you want to test packages in clean environment.

### cmdx gr - Update `registry.yaml`

`cmdx gr` merges `pkgs/**/registry.yaml` and updates [registry.yaml](https://github.com/aquaproj/aqua-registry/blob/main/registry.yaml).
Please don't edit [registry.yaml](https://github.com/aquaproj/aqua-registry/blob/main/registry.yaml) directly.
When you edit `pkgs/**/registry.yaml`, please run `cmdx gr` to reflect the update to `registry.yaml` in the repository.

### cmdx con - Connect to a container

`cmdx con [<os>] [<arch>]` connect to a given container.
`cmdx s` and `cmdx t` tests packages in containers.
`cmdx con` is useful to look into the trouble in containers.
By default, `<os>` is `linux` and `<arch>` is CPU architecture of your machine.

## How to add a package

[Requirements](#requirements), [Set up](#set-up)

1. Scaffold configuration: `cmdx s <package name>`

:::caution
`cmdx s` creates a commit, but please don't edit the commit by `git commit --amend`, `git rebase`, or somehow.
`cmdx s` creates a commit to distinguish scaffolded code from manual changes.
Please add new commits if you update code.
:::

:::caution
Sometimes `cmdx s <package name>` would fail, but this is expected.
In this case, please check the error message and fix `pkgs/<package name>/{pkg.yaml,registry.yaml}`.
Please check [Troubleshooting](/docs/trouble-shooting) too.
If you can't figure out how to fix, please open a pull request and ask us for help.
:::

2. Fix generated files `pkgs/<package name>/{pkg.yaml,registry.yaml}` if necessary
2. Run test: `cmdx t [<package name>]`
2. Update registry.yaml: `cmdx gr`
2. Commit `registry.yaml` and `pkgs/<package name>/{pkg.yaml,registry.yaml`
2. Repeat the step 2 ~ 5 until packages are installed properly
2. Create a pull request
2. (Optional) Stop containers: `cmdx stop`

:::info
We removed `cmdx new` from the guide.
You can still use `cmdx new`, but if you have any trouble with `cmdx new`, you can create a pull request without `cmdx new`.
[Please see the changelog for details.](changelog.md#why-did-we-remove-cmdx-new-from-the-guide)
:::

### Use `cmdx s` definitely

We don't accept pull requests not following this guide.
Especially, we don't accept pull requests not using `cmdx s`.
Standard Registry must support not only the latest version but also almost all versions and [various platforms](#supported-os-and-cpu-architecture).
Many tools have so many versions that people can't check all of them manually.
So we can't trust the code not using `cmdx s`.
`cmdx s` checks all GitHub Releases and generates code supporting all of them (Strictly speaking, if there are too many GitHub Releases we have to restrict the number of GitHub Releases, though `cmdx s` can still check over 200 versions).
`cmdx s` generates much better code than us.

`cmdx s` isn't perfect and sometimes `cmdx s` causes errors and generates invalid code.
Then you have to fix the code according to the error message.
`cmdx s` supports only `github_release` type packages, so for other package types you have to fix the code.
Even if so, you must still use `cmdx s`.
`cmdx s` guarantees the quality of code.

### :bulb: How to ignore some assets and versions

You can ignore some assets and versions to scaffold better configuration files.

:::caution
Be careful to use this feature as it may exclude assets and versions unexpectedly.
Especially, `all_assets_filter` may exclude assets such as checksum files.
We recommend to scaffold codes without this feature first.
Then if `cmdx s` can't scaffold good codes due to some noisy versions or assets, you should re-scaffold code using this feature.
About `all_assets_filter`, we recommend specifying patterns to exclude assets (deny list) rather than specifying patterns to include assets (allow list).

e.g.

```yaml
all_assets_filter: not (Asset contains "static")
```
:::

1. Create `aqua-generate-registry.yaml` by `aqua gr --init` command:

```sh
aqua gr --init <package name>
```

2. Edit `aqua-generate-registry.yaml`:

Example 1. Filter assets:

```yaml
name: argoproj/argo-rollouts
all_assets_filter: not ((Asset matches "rollouts-controller") or (Asset matches "rollout-controller"))
```

Example 2. Filter versions by `version_prefix`:

```yaml
name: grpc/grpc-go/protoc-gen-go-grpc
version_prefix: cmd/protoc-gen-go-grpc/
```

Example 3. Filter versions by `version_filter`:

```yaml
name: crate-ci/typos
version_filter: not (Version startsWith "varcon-")
```

3. Run `cmdx s` with `aqua-generate-registry.yaml`

```sh
cmdx s -c aqua-generate-registry.yaml
```

### :bulb: Set a GitHub Access token to avoid GitHub API rate limiting

If you face GitHub API rate limiting, please set the GitHub Access token with environment variable `GITHUB_TOKEN` or `AQUA_GITHUB_TOKEN`.

e.g.

```sh
export GITHUB_TOKEN=<YOUR PERSONAL ACCESS TOKEN>
```

### How to execute a package in your machine during development

There are several ways

1. Execute a package in linux containers via `cmdx con`
1. Import `pkgs/<package>/pkg.yaml` in [aqua.yaml](https://github.com/aquaproj/aqua-registry/blob/main/aqua.yaml)
1. Add [aqua-all.yaml](https://github.com/aquaproj/aqua-registry/blob/main/aqua-all.yaml) in `$AQUA_GLOBAL_CONFIG`

#### 1. Execute a package in linux containers via `cmdx con`

```console
$ cmdx con
+ bash scripts/connect.sh
[INFO] Connecting to the container aqua-registry (linux/arm64)
```

Then you can execute a package in the container.

#### 2. Import `pkgs/<package>/pkg.yaml` in aqua.yaml

```yaml
packages:
  # ...
  - import: pkgs/<package>/pkg.yaml
```

Please don't commit this change.

You need to run `aqua policy allow` to use the local registry.

```sh
aqua policy allow
```

Then you can execute the package.

#### 3. Add aqua-all.yaml in `$AQUA_GLOBAL_CONFIG`

```sh
export AQUA_GLOBAL_CONFIG=$PWD/aqua-all.yaml:$AQUA_GLOBAL_CONFIG
```

You need to run `aqua policy allow` to use the local registry.

```sh
aqua policy allow
```

Then you can execute all packages.

## Supported OS and CPU Architecture

Please consider the following OS and CPU Architecture.

- OS
  - windows
  - darwin
  - linux
- CPU Architecture
  - amd64
  - arm64

We test the registry in CI on the above environments by GitHub Actions' build matrix.

## Test multiple versions

If the package has the field [version_overrides](/docs/reference/registry-config/version-overrides),
please add not only the latest version but also old versions in `pkg.yaml` to test if old versions can be installed properly.

e.g. [pkg.yaml](https://github.com/aquaproj/aqua-registry/blob/main/pkgs/scaleway/scaleway-cli/pkg.yaml) [registry.yaml](https://github.com/aquaproj/aqua-registry/blob/main/pkgs/scaleway/scaleway-cli/registry.yaml)

```yaml
packages:
  - name: scaleway/scaleway-cli@v2.12.0
  - name: scaleway/scaleway-cli
    version: v2.4.0
```

:warning: Don't use the short syntax `<package name>@<version>` for the old version to prevent Renovate from updating the old version.

:thumbsdown:

```yaml
packages:
  - name: scaleway/scaleway-cli@v2.12.0
  - name: scaleway/scaleway-cli@v2.12.0
```

## What's pkgs/**/pkg.yaml for?

`pkgs/**/pkg.yaml` are test data.
`pkgs/**/pkg.yaml` are used to test if packages can be installed properly.

Note that `pkgs/**/pkg.yaml` aren't lists of available versions.
You can install any versions not listed in `pkgs/**/pkg.yaml`.

## Trouble shooting

### `cmdx new` fails to push a commit to the origin

:::info
We removed `cmdx new` from the guide.
You can still use `cmdx new`, but if you have any trouble with `cmdx new`, you can create a pull request without `cmdx new`.
[Please see the changelog for details.](changelog.md#why-did-we-remove-cmdx-new-from-the-guide)
:::

If `cmdx new` can't push a commit to a remote branch, please confirm if `origin` is not the upstream `aquaproj/aqua-registry` but your fork.
If `origin` is not your fork, please change it to your fork.

e.g. Fail to push a commit

```console
$ cmdx new pre-commit/pre-commit
# ...
+ git push origin feat/pre-commit/pre-commit
remote: Permission to aquaproj/aqua-registry.git denied to ***.
fatal: unable to access 'https://github.com/aquaproj/aqua-registry/': The requested URL returned error: 403
```

1. [If you haven't forked aquaproj/aqua-registry, please fork it](https://github.com/aquaproj/aqua-registry/fork).
2. Check remote repositories.

```sh
git remote -v
```

3. Please fix `origin`.

```sh
git remote set-url origin https://github.com/<your fork>
```

4. Please set `upstream` if necessary.

```sh
git remote add upstream https://github.com/aquaproj/aqua-registry
```
