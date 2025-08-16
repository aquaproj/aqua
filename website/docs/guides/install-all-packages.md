---
sidebar_position: 600
---

# Install Standard Registry's all packages very quickly

You can make Standard Registry's all packages available in your laptop very quickly.
If you have to install all packages at the same time, the experience would be too bad.
It would consume many disk and it would take a long time.
But no worries.
By Lazy Install, packages aren't installed until they are really needed.

1. Check out this repository
1. Add [aqua-all.yaml](https://github.com/aquaproj/aqua-registry/blob/main/aqua-all.yaml) to the environment variable `AQUA_GLOBAL_CONFIG`
1. Run `aqua i -l -a`

```console
$ git clone https://github.com/aquaproj/aqua-registry
$ export AQUA_GLOBAL_CONFIG="$PWD/aqua-registry/aqua-all.yaml:$AQUA_GLOBAL_CONFIG"
$ aqua i -l -a
```

`aqua i -l -a` would finish immediately, because it only creates symbolic links.

By Setting up cron to checkout the repository and run `aqua i -l -a` periodically, you can update packages automatically.

If you want to change some packages' version, please override them by the other configuration file.

```console
$ export AQUA_GLOBAL_CONFIG="<Other aqua.yaml>:$PWD/aqua-all.yaml:$AQUA_GLOBAL_CONFIG"
```
