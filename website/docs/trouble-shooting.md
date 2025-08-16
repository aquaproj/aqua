---
sidebar_position: 700
---

# Trouble Shooting

## Use the latest aqua and aqua-registry

If aqua and aqua-registry are old, please update them.

## Set AQUA_LOG_LEVEL=debug

When aqua doesn't work, the environment variable `AQUA_LOG_LEVEL` is helpful for the debug.

For example,

```console
$ AQUA_LOG_LEVEL=debug kind --help
DEBU[0000] CLI args                                      config= log_level=debug program=aqua
DEBU[0000] install the package                           package_name=kubernetes-sigs/kind package_version=v0.11.1 program=aqua registry=standard
DEBU[0000] check if the package is already installed     package_name=kubernetes-sigs/kind package_version=v0.11.1 program=aqua registry=standard
DEBU[0000] check the permission                          file_name=kind
DEBU[0000] execute the command                           exe_path=/Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64 program=aqua
DEBU[0000] command was executed but it failed            error="fork/exec /Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64: exec format error" exe_path=/Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64 exit_code=-1 program=aqua
DEBU[0000] command failed                                error="fork/exec /Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64: exec format error" exit_code=-1 program=aqua
```

In the above case, kind is already installed but the `exec format error` occurred. I reinstalled kind then the problem has been solved.

```console
# uninstall kind
$ rm -R /Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64
# kind is reinstalled by Lazy Install
$ kind --help
```

## Tracing and CPU Profiling

If you encounter any performance issue, please see [Tracing and CPU Profiling](/docs/reference/config/trace-profile).

## check file_src is correct

If aqua outputs the warning or error `check file_src is correct`,
this means that the asset was downloaded and unarchived but the executable file wasn't found.
Probably this is the problem of the Registry, so please update the Registry.

e.g.

```console
$ aqua i
WARN[0000] check file_src is correct                     aqua_version=1.15.1 env=darwin/arm64 error="exe_path isn't found: stat /Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/neovim/neovim/v0.7.2/nvim-macos.tar.gz/nvim-osx64/bin/nvim: no such file or directory" exe_path=/Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/neovim/neovim/v0.7.2/nvim-macos.tar.gz/nvim-osx64/bin/nvim file_name=nvim package=neovim/neovim package_name=neovim/neovim package_version=v0.7.2 program=aqua registry=standard
```

In this case, the file `/Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/neovim/neovim/v0.7.2/nvim-macos.tar.gz/nvim-osx64/bin/nvim` wasn't found.

Please check the correct path.

```console
$ ls /Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/neovim/neovim/v0.7.2/nvim-macos.tar.gz
nvim-macos
```

The directory name was not `nvim-osx` but `nvim-macos`.

```console
$ ls /Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/neovim/neovim/v0.7.2/nvim-macos.tar.gz/nvim-macos/bin 
nvim
```

So we fixed the Standard Registry.

* https://github.com/aquaproj/aqua/issues/970#issuecomment-1171726476
* https://github.com/aquaproj/aqua-registry/pull/4419

## the asset isn't found

If aqua outputs the error `the asset isn't found`, the following are possible causes.

* The release doesn't exist
  * Please fix the version
* GitHub Releases has no assets yet
  * Please wait for uploading assets. This isn't a problem of aqua
  * Renovate's [minimumReleaseAge](https://docs.renovatebot.com/configuration-options/#minimumreleaseage) may be useful to decrease this kind of issues
* Assets for only specific pair of OS and Arch aren't uploaded
  * Maybe you can request to release assets to the tool owner
  * e.g. https://github.com/gsamokovarov/jump/issues/72
* Asset name format was changed from the specific version
  * You have to fix the Registry
  * In case of the Standard Registry, please create an issue or a pull request

## GitHub API Rate Limit, 403 error

Please set GitHub Access Token to the environment variable `GITHUB_TOKEN` or `AQUA_GITHUB_TOKEN`.

## Packages aren't installed

Maybe your OS and Arch isn't supported by the package's [supported_envs](/docs/reference/registry-config/supported-envs).
Please check the Registry Configuration.

## Command is not found

If `command -v <command>` exits with non zero, the following are possible causes.

* `AQUA_ROOT_DIR/bin` isn't added to the environment variable `PATH`
  * e.g. `$ export PATH=$HOME/.local/share/aquaproj-aqua/bin:$PATH`
* the symbolic link isn't created in `AQUA_ROOT_DIR/bin`
  * Please run `aqua i -l`
* the command name is wrong

You can check the package's command names by `aqua g` command.
For example, the command name of the package `cli/cli` is `gh`.

```console
$ aqua g
```

```
  docker-slim/docker-slim [docker-slim, docker-slim-sensor]
  corneliusweig/rakkess/access-matrix [kubectl-access_matrix]
  CircleCI-Public/circleci-cli [circleci]
> cli/cli [gh]: github
  4/660
> cli/cli
```

If the symbolic link isn't created by `aqua i -l`, the following are possible causes.

* Your OS and Arch isn't supported by the package's [supported_envs](/docs/reference/registry-config/supported-envs)
* the package isn't added in configuration files

aqua finds the configuration files and packages according to the rule.

* [Configuration file path | Tutorial](/docs/tutorial/config-path)
* [Configuration file paths | Reference](/docs/reference/config#configuration-file-path)

Please check configuration files and your current directory.

If `command -v <command>` exits with zero but command can't executed by the error `error="command is not found"`, the following are possible causes.

e.g.

```console
$ gh version
FATA[0000] aqua failed                                   aqua_version=1.19.2 error="command is not found" exe_name=gh program=aqua
```

aqua finds the configuration files and packages according to the rule.

* [Configuration file path | Tutorial](/docs/tutorial/config-path)
* [Configuration file paths | Reference](/docs/reference/config#configuration-file-path)

Please check configuration files and your current directory.

## The tool X doesn't work well

When the tool X managed by aqua is executed, X is intermediated by aqua-proxy and aqua.
Please see [here](/docs/reference/execve-2) too.
Due to this intermediation, there are cases that some tools don't work well.

There is a workaround that you can try when you face the issue.
The workaround is to execute the tool directly by executing `aqua which X` and getting the absolute path.

For example, when we tried [LunarVim](https://www.lunarvim.org/) we faced the issue that LunarVim didn't start.
The issue occurred as we managed NeoVim with aqua.
LunarVim executed NeoVim as the following.

```sh
# $HOME/.local/bin/lvim
exec -a "$NVIM_APPNAME" nvim -u "$LUNARVIM_BASE_DIR/init.lua" "$@"
```

To resolve the issue, we replaced `nvim` with `"$(aqua which nvim)"`

```sh
# $HOME/.local/bin/lvim
exec -a "$NVIM_APPNAME" "$(aqua which nvim)" -u "$LUNARVIM_BASE_DIR/init.lua" "$@"
```

Then the issue was solved and we could start LunarVim!
