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

## Disable tracking

`aqua >= v2.63.0`

If the environment variable `AQUA_DISABLE_TRACKING` is `true`, aqua doesn't record packages' last used date times.

```sh
export AQUA_DISABLE_TRACKING=true
```

This is useful if `$AQUA_ROOT_DIR` is read only.
For instance, an administrator installs packages in a shared directory and other users only execute them.
In that case aqua fails to record last used date times and outputs warning logs every time packages are executed.
`AQUA_DISABLE_TRACKING` suppresses those warning logs.

`aqua vacuum` and `aqua vacuum --init` fail while `AQUA_DISABLE_TRACKING` is set.

```console
$ aqua vacuum
ERR aqua failed doc="https://aquaproj.github.io/docs/reference/codes/007" error="the vacuum command isn't available. Tracking is disabled" program=aqua
```

They fail rather than doing nothing because last used date times get stale without tracking.
`aqua vacuum` would judge packages in use as unused and remove them.

If you want to run `aqua vacuum` again, please unset `AQUA_DISABLE_TRACKING` and run `aqua vacuum --init` to record the current date time as the last used date time of installed packages.

Note that `aqua rm` still removes a timestamp file of a removed package even if `AQUA_DISABLE_TRACKING` is `true`, so that stale files aren't left behind.
