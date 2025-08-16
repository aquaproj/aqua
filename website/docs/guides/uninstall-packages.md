---
sidebar_position: 500
---

# Uninstall Packages

:::info
See also [Remove unused packages (Vacuum)](./vacuum.md)
:::

aqua >= [v2.11.0](https://github.com/aquaproj/aqua/releases/tag/v2.11.0) [#538](https://github.com/aquaproj/aqua/issues/538) [#2248](https://github.com/orgs/aquaproj/discussions/2248) [#2249](https://github.com/aquaproj/aqua/pull/2249)

:::caution
`aqua rm` command doesn't remove packages from `aqua.yaml`.
:::

```console
$ aqua rm --all [-a] # Uninstall all packages
```

```console
$ aqua rm [<registry name>,]<package name> [...] # Uninstall packages
```

```console
$ aqua rm <command name> [...] # Uninstall packages having given commands
```

```console
$ aqua rm -i # Select packages interactively with a fuzzy finder
```

e.g.

```console
$ aqua rm cli/cli direnv/direnv
```

## -mode option

aqua >= v2.32.0 [#3075](https://github.com/aquaproj/aqua/pull/3075)

By default, `aqua remove` command removes only packages from the `pkgs` directory and doesn't remove links from the `bin` directory.
You can change this behaviour by specifying the `-mode` flag.
The value of `-mode` is a string containing characters `l` and `p`.
The order of the characters doesn't matter.

```sh
aqua rm -m l cli/cli # Remove only links
aqua rm -m pl cli/cli # Remove links and packages
```

You can also configure the mode by the environment variable `AQUA_REMOVE_MODE`, so you can change the default behaviour of `aqua remove` command by setting `AQUA_REMOVE_MODE` in your shell setting such as `.bashrc`.

```sh
export AQUA_REMOVE_MODE=pl
```

## Limitation

:::info
As of [aqua v2.33.0](https://github.com/aquaproj/aqua/releases/tag/v2.33.0), you can uninstall `go_install` and `http` packages too.
:::

1. The fuzzy finder shows all packages, which include not installed packages.
