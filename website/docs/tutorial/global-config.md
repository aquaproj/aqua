---
sidebar_position: 600
---

# Install tools globally

aqua finds the configuration files from the current directory to the root directory.

```console
$ pwd
/tmp
$ gh version
FATA[0000] aqua failed                                   aqua_version=1.19.2 error="command is not found" exe_name=gh program=aqua
```

If you want to install tools regardless the current directory,
let's add the global configuration.
Create a global configuration file and add the path to the environment variable `AQUA_GLOBAL_CONFIG`.
You can change the global configuration file path freely.

```console
$ mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/aquaproj-aqua"
$ vi "${XDG_CONFIG_HOME:-$HOME/.config}/aquaproj-aqua/aqua.yaml"
$ export AQUA_GLOBAL_CONFIG=${AQUA_GLOBAL_CONFIG:-}:${XDG_CONFIG_HOME:-$HOME/.config}/aquaproj-aqua/aqua.yaml
```

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry

packages:
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
```

```console
$ gh version
gh version 2.2.0 (2021-10-25)
https://github.com/cli/cli/releases/tag/v2.2.0
```

## `aqua i` ignores global configuration by default

:::caution
`aqua i` ignores global configuration by default.
To install tools of global configuration by `aqua i`, please set the `-a` option.

```console
$ aqua i -a
```
:::

## See also

- [Share aqua configuration for teams and organizations](/docs/guides/team-config)
- [Install Standard Registry's all packages very quickly](/docs/guides/install-all-packages)
