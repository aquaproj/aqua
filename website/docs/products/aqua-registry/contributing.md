---
sidebar_position: 100
---

# Contributing

How to contribute to Standard Registry. https://github.com/aquaproj/aqua-registry

## See also

- :star: [OSS Contribution Guide](https://github.com/suzuki-shunsuke/oss-contribution-guide)
- :star: [Registry Style Guide](/docs/develop-registry/registry-style-guide)
- [Registry Configuration](/docs/reference/registry-config/)

## Changelog of document and development workflow

[change log](changelog.md).

- [2025-01-19 Add `Note of Programing Language Support`](#note-of-programing-language-support)
- [2024-12-14 Remove `cmdx new` from the guide](changelog.md#2024-12-14)
- [2024-05-24 The behaviour of `cmdx s`, `cmdx t`, and `cmdx new` were changed.](changelog.md#2024-05-24)

## Prerequisites

### Commit Signing

All commits of pull requests must be signed.
Please see [the document](https://github.com/suzuki-shunsuke/oss-contribution-guide/blob/main/docs/commit-signing.md).

### Set up

1. [Please fork aquaproj/aqua-registry](https://github.com/aquaproj/aqua-registry/fork).
1. Checkout the repository
2. Run `aqua i -l` in the repository root directory to install tools which are required for contribution.

```sh
aqua i -l
```

### GitHub Access Token

Development tools execute GitHub API to get lists of GitHub Releases and assets.
It works without an access token, but the possibility of hitting API rate limits increases.
Hitting API rate limits can prevent normal code generation or cause tests to fail.
You can pass an access token through environment variables `GITHUB_TOKEN` or `AQUA_GITHUB_TOKEN`.
If these environment variables are not set, it will try to get an access token using the `gh auth token` command.
No special permissions are needed as it only reads public repository resources.

## Should you create an Issue before sending a Pull Request?

Basically, you don't have to create an Issue before sending a Pull Request.
But if the pull request requires the discussion before reviewing, you have to create an Issue in advance.

For example, you don't have to create an Issue in the following cases.

- Add a package
- Fix a typo

On the other hand, for example if you want to change the directory structure in `pkgs` or the workflow adding a package,
you have to create an Issue and describe what is changed and why the change is needed.

## Structure of aqua-registry

Package-related code is located in the `pkgs/<package name>` directory of https://github.com/aquaproj/aqua-registry.
e.g. [cli/cli](https://github.com/aquaproj/aqua-registry/tree/main/pkgs/cli/cli)
Each package directory contains the following files:

- pkg.yaml: List of versions installed during testing. This is essentially test data
- registry.yaml: Configuration. Each tool's registry.yaml is merged to generate the repository root registry.yaml
- scaffold.yaml: Optional. Configuration file for commands that auto-generate pkg.yaml and registry.yaml. Required when you want to change the auto-generation behavior

:::note
pkg.yaml is just test data. You can install versions not included in this file.
:::

There is also a registry.yaml at the repository root, which is a huge YAML file merging all registry.yaml files under `pkgs`.
When specifying Standard Registry in aqua.yaml, this repository root registry.yaml is referenced.
To modify the repository root registry.yaml, modify the registry.yaml under pkgs and run the `cmdx gr` command.

### registry.yaml Documentation

Please refer to [Registry Config](/docs/reference/registry-config/).
There is also a [JSON Schema](https://github.com/aquaproj/aqua/blob/main/json-schema/registry.json).
The registry.yaml files under pkgs have JSON Schema comments, so VSCode and similar editors can provide auto-completion.

Additionally, there are abundant examples under pkgs in aqua-registry.
By grepping here, you can see how much each configuration item is used and how to write them for reference.

```console
# Search with slsa_provenance
$ git grep -l slsa_provenance pkgs  
pkgs/Zxilly/go-size-analyzer/registry.yaml
pkgs/aquaproj/aqua-registry-updater/registry.yaml
pkgs/aquaproj/example-go-slsa-provenance/registry.yaml
...
```

### What's pkgs/**/pkg.yaml for?

`pkgs/**/pkg.yaml` are test data.
`pkgs/**/pkg.yaml` are used to test if packages can be installed properly.

Note that `pkgs/**/pkg.yaml` aren't lists of available versions.
You can install any versions not listed in `pkgs/**/pkg.yaml`.

## Development Tools

:::info
Unfortunately, the current development tools depend on Shell Scripts and are unlikely to work on Windows (though they probably work on WSL).
[There is an issue to rewrite them in Go.](https://github.com/aquaproj/aqua-registry/issues/32699)
:::

CLIs for developing aqua-registry are provided.
These can be installed with aqua.
Check out aqua-registry and run `aqua i -l`.

```sh
aqua i -l
```

We use a task runner called [cmdx](https://github.com/suzuki-shunsuke/cmdx).
You can check tasks with `cmdx help`.

```sh
cmdx help
```

Using development tools, you can generate files for each package (pkg.yaml, registry.yaml, scaffold.yaml) and test tool installation in containers.
Tests are performed in containers using the docker command, so you need the docker command and a compatible container engine.
Docker Desktop would work fine of course.
We install [docker/cli](https://github.com/docker/cli) and [abiosoft/colima](https://github.com/abiosoft/colima) with aqua.
Please confirm that the docker command works.

```sh
docker version
```

### cmdx Commands Overview

#### cmdx s - Scaffold configuration and test it in containers

`cmdx s <package name>` generates a configuration file `pkgs/<package name>/registry.yaml` and a test data file `pkgs/<package name>/pkg.yaml`, and tests them in containers.
It gets data from GitHub Releases by GitHub API.
By default, it gets all releases, so it takes a bit long time if the repository has a lot of releases.

**Important**: `cmdx s` is required when adding new packages. We don't accept pull requests not using `cmdx s`.
Standard Registry must support not only the latest version but also almost all versions and [various platforms](#supported-os-and-cpu-architecture).
`cmdx s` checks all GitHub Releases and generates code supporting all of them (can check over 200 versions).
`cmdx s` guarantees the quality of code.

For package types other than github_release, specify `-l 1`:

```sh
cmdx s -l 1 "<package name>"
```

:::caution
`cmdx s` creates a commit, but please don't edit the commit by `git commit --amend`, `git rebase`, or somehow.
`cmdx s` creates a commit to distinguish scaffolded code from manual changes.
Please add new commits if you update code.
:::

#### cmdx t - Test a package in containers

`cmdx t [<package name>]` tests a package in containers.
If the branch name is `feat/<package name>`, you can omit the argument `<package name>`.
`cmdx t` copies files `pkgs/<package name>/{pkg.yaml,registry.yaml}` in containers and test them.
If the test succeeds, `registry.yaml` is updated.

#### cmdx gr - Update `registry.yaml`

`cmdx gr` merges `pkgs/**/registry.yaml` and updates [registry.yaml](https://github.com/aquaproj/aqua-registry/blob/main/registry.yaml).
Please don't edit [registry.yaml](https://github.com/aquaproj/aqua-registry/blob/main/registry.yaml) directly.
When you edit `pkgs/**/registry.yaml`, please run `cmdx gr` to reflect the update to `registry.yaml` in the repository.

#### cmdx rm - Remove containers

`cmdx rm` removes containers.
`cmdx s` and `cmdx t` reuse containers, but if you want to test packages in clean environment, you can do it by removing containers.

#### cmdx rmp - Remove an installed package from containers

`cmdx rmp [<package name>]` removes an installed package from containers.
If the branch name is `feat/<package name>`, you can omit the argument `<package name>`.
It runs `aqua rm <package name>` and removes `aqua-checksums.json` in containers.
This task is useful when you want to test packages in clean environment.

#### cmdx con - Connect to a container

`cmdx con [<os>] [<arch>]` connect to a given container.
`cmdx s` and `cmdx t` tests packages in containers.
`cmdx con` is useful to look into the trouble in containers.
By default, `<os>` is `linux` and `<arch>` is CPU architecture of your machine.

## Working with Packages

### Tool Naming Convention

To avoid name conflicts, tool names must include `/` (namespace-like meaning).

- NG: `terraform`
- OK: `hashicorp/terraform`

If the tool code is managed on GitHub, match the repository name.
If multiple tools are managed in that repository, change the name for each tool.

e.g. [winebarrel/cronplan](https://github.com/winebarrel/cronplan)

- `winebarrel/cronplan/cronmatch`
- `winebarrel/cronplan/cronplan`
- `winebarrel/cronplan/cronviz`

Packages hosted outside GitHub should have naming that distinguishes them from GitHub.
`cargo` packages become [crates.io/{crate name}](https://github.com/aquaproj/aqua-registry/tree/main/pkgs/crates.io).
Platforms other than GitHub like GitLab are not actively supported, but some are supported as http type packages.
[GitLab uses `gitlab.com/<repository name>`.](https://github.com/aquaproj/aqua-registry/tree/main/pkgs/gitlab.com)

### Adding New Tools

When submitting a Pull Request to add a new tool, there's no need to create an Issue.

Run `cmdx s` to auto-generate code:

```sh
cmdx s "<tool name>"
```

e.g.

```sh
cmdx s cli/cli
```

For package types other than github_release, specify `-l 1`:

```sh
cmdx s -l 1 "<package name>"
```

cmdx s generates a branch `feat/<package name>`, code, and commit, and tests using containers.

:::info
This command may sometimes fail tests and output a large amount of error messages, but don't be overwhelmed by those error messages.
Test failures are expected.
:::

#### Customizing Generation with Configuration File

Sometimes `cmdx s` generation doesn't work in one go.
For github_release packages, `cmdx s` gets lists of GitHub Releases and assets via GitHub API and auto-generates configuration based on that.
However, sometimes you need to exclude specific versions or assets.

In such cases, follow these steps:

1. Generate a template configuration file `aqua-generate-registry.yaml` for `cmdx s` with `aqua gr -init <package name>`
2. Modify the configuration file `aqua-generate-registry.yaml`
3. Generate code with `cmdx s -c "<configuration file>" "<package name>"`

You can configure the following:

- `version_filter`: Versions not matching this condition are excluded
- `version_prefix`: Versions without this prefix are excluded
- `all_assets_filter`: Assets not matching this condition are excluded

:::caution
Be careful to use this feature as it may exclude assets and versions unexpectedly.
Especially, `all_assets_filter` may exclude assets such as checksum files.
We recommend to scaffold codes without this feature first.
Then if `cmdx s` can't scaffold good codes due to some noisy versions or assets, you should re-scaffold code using this feature.
About `all_assets_filter`, we recommend specifying patterns to exclude assets (deny list) rather than specifying patterns to include assets (allow list).

Note that `version_filter` is not a feature for dropping support for old versions.
`version_constraint`, `no_asset`, and `error_message` are used for dropping support for old versions.
:::

Example configurations:

```yaml
# Filter assets
name: argoproj/argo-rollouts
all_assets_filter: not ((Asset matches "rollouts-controller") or (Asset matches "rollout-controller"))
```

```yaml
# Filter versions by prefix
name: grpc/grpc-go/protoc-gen-go-grpc
version_prefix: cmd/protoc-gen-go-grpc/
```

```yaml
# Filter versions by expression
name: crate-ci/typos
version_filter: not (Version startsWith "varcon-")
```

#### Retrying Generation

As mentioned earlier, code generation with `cmdx s` doesn't always work on the first try.
Sometimes you need to repeat it several times.

1. Generate code without configuration file `cmdx s`
2. Check the generated code, and if extra versions or assets are included, delete the generated branch

```sh
git checkout main
git branch -D "feat/<package name>"
```

3. Generate configuration file `aqua gr -init`
4. Modify configuration file and generate code `cmdx s`
5. Repeat 2, 4 until extra versions and assets are excluded

### Modifying Existing Packages

When modifying existing packages, you need to modify code under `pkgs/<package name>`.
There are several modification methods:

1. Manually modify the code
2. Regenerate the code from scratch with commands
3. Auto-generate code for the latest version and manually modify based on that

Which method to use depends on the state of the original code.
Code auto-generation has been improved many times.
Therefore, there is low-quality code generated before improvements.
Such code may be better regenerated from scratch rather than manually fixed.

One characteristic to identify if code is old is how `version_constraint` and `version_overrides` are written.
In the new style, it basically looks like this:

```yaml
  version_constraint: "false" # Root version_constraint is "false"
  version_overrides:
    - version_constraint: semver("<= 0.1.0") # Version constraints use <, <= not >, >= (basically <=)
      # ...
    # ...
    - version_constraint: "true" # End with "true" for latest version configuration
      # ...
```

In the old style, `version_overrides` is often not defined.
In this case, it's likely better to regenerate from scratch.
However, auto-generation doesn't support package types other than `github_release` or `cargo`, so manual modification will be necessary.

Also, [aliases](https://aquaproj.github.io/docs/reference/registry-config/aliases) and [files](https://aquaproj.github.io/docs/reference/registry-config/files) cannot be auto-generated, so you need to modify the auto-generated code referring to the original code.

To generate code for the latest version only:

```sh
aqua gr -l 1 "<package name>"
```

Fix this and add it to the end of `version_overrides` in the original code and modify version_constraint.

### Manual Modifications

When manual modification is necessary, you'll need to look at error messages and fix appropriately.
If installation of multiple versions is failing and the log is hard to read, it's good to comment out some versions in pkg.yaml and tackle problems one by one.
After modification, run `cmdx t` to confirm it can be installed correctly.

#### Common Issues and Solutions

##### When Configuration Needs to Change for Specific Versions

You can change configuration by version using [version_overrides and version_constraint](/docs/reference/registry-config/version-overrides).

##### When Configuration Needs to Change for Specific OS/Arch

You can change configuration by OS/Arch with [overrides](/docs/reference/registry-config/overrides).

##### When Version Cannot Be Found

Sometimes a released version is deleted and disappears.
In that case, delete that version from pkg.yaml.
And delete configuration related to that version from registry.yaml (if any).
However, [no_asset](/docs/reference/registry-config/no_asset) and [error_message](/docs/reference/registry-config/error_message) don't need to be deleted.

##### When Asset Cannot Be Found

When an asset cannot be found, either the asset name is wrong or the asset hasn't been released.

Running the `cmdx lsa [-r <repository name>] "<version>"` command outputs a list of assets, which is convenient.

```console
$ cmdx lsa -repo suzuki-shunsuke/pinact v3.0.0
+ REPO=${REPO#https://github.com/}
repo=$(bash scripts/get_test_pkg.sh "$REPO")

gh release view --json assets --jq ".assets[].name" -R "$repo" "$VERSION"

multiple.intoto.jsonl
pinact_3.0.0_checksums.txt
pinact_3.0.0_checksums.txt.pem
pinact_3.0.0_checksums.txt.sig
pinact_darwin_amd64.tar.gz
pinact_darwin_arm64.tar.gz
pinact_linux_amd64.tar.gz
pinact_linux_arm64.tar.gz
pinact_windows_amd64.zip
pinact_windows_arm64.zip
```

Common causes when there are no assets:

1. Release is simply delayed. It will be released if you wait
2. CI failed midway and wasn't released
3. CI skipped the release

These are not problems with aqua or aqua-registry.
For example, if such a problem occurs with [suzuki-shunsuke/pinact](https://github.com/suzuki-shunsuke/pinact) and you want to take action, it would be good to create an issue or PR at https://github.com/suzuki-shunsuke/pinact.

It's common for specific os/arch not to be supported.
In that case, you need to exclude that os/arch from `supported_envs`.

If the asset name is wrong, the asset naming convention may have changed from a certain version.
In that case, you need to modify the asset in registry.yaml.

##### When Command Cannot Be Found

When a command cannot be found, the following possibilities exist:

1. Command name is wrong
2. Command name changed
3. Path is wrong
4. Target os/arch is excluded by supported_envs

In these cases, you need to modify the `files` configuration.

```yaml
files:
  - name: <command name>
    src: <relative path to command executable>
```

By default, the command name is the last element when splitting the package name by `/`.
So for `cli/cli` it becomes `cli`, but the actual command name is `gh`, so you need to explicitly specify `files`.

```yaml
files:
  - name: gh
```

Note that even on Windows, `.exe` is not added to the name.

`src` is the relative path where the command executable is located when extracting assets like tarball or zip.
By default, it's the same as `name`.
For gh, since the path is different, you need to specify `src`.

```yaml
    files:
      - name: gh
        src: gh_{{trimV .Version}}_{{.OS}}_{{.Arch}}/bin/gh
```

The auto-generation tool currently cannot auto-generate `files`.
Therefore, manual modification is necessary.

##### Adding Support for Specific OS / Architecture

Sometimes a tool supports new OS/Architecture from a specific version but it's not reflected in registry.yaml and remains uninstallable.
In that case, you need to add that OS/Architecture to `supported_envs`.

##### When Checksum Cannot Be Extracted from Checksum File

Please see [the document](/docs/reference/registry-config/checksum).

##### When Checksum Verification Fails

Please see [the document](/docs/reference/registry-config/checksum).

1. Checksum written in checksum file is wrong => Disable checksum

```yaml
checksum:
  enabled: false
```

Or delete the checksum configuration since it's disabled by default.

:::info
The checksum enable/disable setting in registry configuration is just a setting for "whether to download checksum file and get checksum".
Even if this is disabled, if checksum verification is enabled in aqua.yaml, checksum verification will be performed.
In that case, it actually downloads the asset, calculates the checksum, and records it in aqua-checksums.json.
:::

2. Extracting wrong string from checksum file

Modify extraction parameters or disable checksum.

3. Wrong checksum algorithm (sha1, sha256, sha512, md5, etc) => Fix the algorithm

##### When cosign Verification Fails

[Please see the document](/docs/reference/registry-config/cosign).

##### When SLSA Provenance Verification Fails

[Please see the document](/docs/reference/registry-config/slsa-provenance).

##### When GitHub Artifact Attestations Verification Fails

[Please see the document](/docs/reference/registry-config/github-artifact-attestations).

`signer_workflow` might be wrong.
If attestations are not generated for a specific version in the first place, delete the github_artifact_attestations configuration.

The github_artifact_attestations configuration cannot be auto-generated currently.
Therefore, when adding a new tool, check if attestations are generated and add the configuration if they are.

##### When Minisign Verification Fails

[Please see the document](/docs/reference/registry-config/minisign).

The minisign configuration might be wrong.
If minisign signing is not performed for a specific version in the first place, delete the minisign configuration.

## Testing

### How to execute a package in your machine during development

There are several ways:

1. Execute a package in linux containers via `cmdx con`
2. Import `pkgs/<package>/pkg.yaml` in [aqua.yaml](https://github.com/aquaproj/aqua-registry/blob/main/aqua.yaml)
3. Add [aqua-all.yaml](https://github.com/aquaproj/aqua-registry/blob/main/aqua-all.yaml) in `$AQUA_GLOBAL_CONFIG`

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

### Test multiple versions

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

## Platform Support

### Supported OS and CPU Architecture

Please consider the following OS and CPU Architecture.

- OS
  - windows
  - darwin
  - linux
- CPU Architecture
  - amd64
  - arm64

We test the registry in CI on the above environments by GitHub Actions' build matrix.

### aqua can't support some tools' plugin mechanism

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

### Note of Programing Language Support

aqua supports several programing languages such as Go and Node.js, but when we support a programing language, we need to be careful about where the programing language installs libraries and commands.

For instance, if the programing language installs commands in the same directory with the programing language itself, aqua can't add them to $PATH, meaning we can't execute them.
aqua doesn't support changing $PATH dynamically (We have no plan to support it as it makes aqua more complicated).
Node.js's `npm i -g` installs the same directory with node by default, so we gave up the support of Node.js before (Now aqua supports Node.js again because we can change the install path by `NPM_CONFIG_PREFIX`).
If the language installs libraries in the same directory with it, the language can't refer installed libraries when we change the version of the language.

So before supporting a programing language, we should consider carefully if it really works well.
Many programing languages have dedicated version managers, so maybe they are more appropriate.

## Troubleshooting

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
