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

## Read only `$AQUA_ROOT_DIR`

Sometimes `$AQUA_ROOT_DIR` is read only for users.
For instance, an administrator installs packages in a shared directory such as `/opt/aqua` and other users only execute them.

In that case aqua fails to record last used date times and outputs a warning log every time packages are executed.

```
WRN update the last used datetime ... error="update the last used datetime: create a package timestamp file: open /opt/aqua/metadata/pkgs/http/get.helm.sh/helm-v4.2.3-linux-amd64.tar.gz/timestamp.txt: permission denied"
```

There are two solutions.
[Granting write permission](#grant-write-permission-to-the-metadata-directory) is better if it's acceptable, because [disabling tracking](#disable-tracking) makes `aqua vacuum` unavailable for everyone including the administrator.

### Grant write permission to the metadata directory

`aqua >= v2.63.0`

aqua records last used date times in `$AQUA_ROOT_DIR/metadata`, which is separated from installed packages in `$AQUA_ROOT_DIR/pkgs`.
So you can allow users to record last used date times while keeping installed packages read only.

```sh
sudo chgrp -R aqua /opt/aqua/metadata
sudo chmod -R g+w /opt/aqua/metadata
sudo find /opt/aqua/metadata -type d -exec chmod g+s {} +
```

Users need to belong to the group and to set `umask` to `0002`, because aqua creates timestamp files and directories with `0664` and `0775` and `umask` removes the group write permission.

```sh
umask 0002
```

The administrator needs to set `umask` too when installing packages.
aqua records the last used date time when it installs a package, so timestamp files of newly installed packages are created by the administrator.
Users can't update them if they aren't group writable.

Warning logs go away, and the administrator can run `aqua vacuum` based on the actual usage of all users.

:::info
aqua < v2.63.0 creates timestamp files with `0644`, so a timestamp file created by a user can't be updated by other users.
:::

### Disable tracking

`aqua >= v2.63.0`

If you can't grant write permission, for instance because `$AQUA_ROOT_DIR` is on a read only file system, you can disable the tracking of last used date times.
If the environment variable `AQUA_DISABLE_TRACKING` is `true`, aqua doesn't record packages' last used date times, so the warning logs go away.

```sh
export AQUA_DISABLE_TRACKING=true
```

:::caution
Last used date times aren't recorded while `AQUA_DISABLE_TRACKING` is set, so the administrator can't remove unused packages based on the actual usage either.
:::

`aqua vacuum` and `aqua vacuum --init` fail while `AQUA_DISABLE_TRACKING` is set.

```console
$ aqua vacuum
ERR aqua failed doc="https://aquaproj.github.io/docs/reference/codes/007" error="the vacuum command isn't available. Tracking is disabled" program=aqua
```

They fail rather than doing nothing because last used date times get stale without tracking.
`aqua vacuum` would judge packages in use as unused and remove them.

If you want to run `aqua vacuum` again, please unset `AQUA_DISABLE_TRACKING` and run `aqua vacuum --init` to record the current date time as the last used date time of installed packages.

Note that `aqua rm` still removes a timestamp file of a removed package even if `AQUA_DISABLE_TRACKING` is `true`, so that stale files aren't left behind.
