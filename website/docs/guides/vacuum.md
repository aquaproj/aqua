---
sidebar_position: 320
---

# Remove unused packages (Vacuum)

:::info
See also [Uninstall packages](./uninstall-packages.md)
:::

[v2.43.0](https://github.com/aquaproj/aqua/releases/tag/v2.43.0) [#3467](https://github.com/aquaproj/aqua/pull/3467)

You can remove unused packages by `aqua vacuum` command, which is useful to save storage and keep your machine clean.

```sh
aqua vacuum
```

This command removes installed packages which haven't been used for over the expiration days.
The default expiration days is 60, but you can change it by the environment variable `$AQUA_VACUUM_DAYS` or the command line option `aqua vacuum -days <expiration days>`.

e.g.

```sh
export AQUA_VACUUM_DAYS=90
```

```sh
aqua vacuum -d 30
```

:::info
aqua vacuum command doesn't remove links from the bin directory and doesn't remove packages from aqua.yaml
:::

As of aqua v2.43.0, aqua records packages' last used date times.
Date times are updated when packages are installed or executed.
Packages installed by aqua v2.42.2 or older don't have records of last used date times, so aqua can't remove them.
To solve the problem, `aqua vacuum --init` is available.

```sh
aqua vacuum --init
```

`aqua vacuum --init` searches installed packages from aqua.yaml including `$AQUA_GLOBAL_CONFIG` and records the current date time as the last used date time of those packages if their last used date times aren't recorded.

`aqua vacuum --init` can't record date times of install packages which are not found in aqua.yaml.
If you want to record their date times, you need to remove them by `aqua rm` command and re-install them.
