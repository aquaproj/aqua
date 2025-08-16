---
sidebar_position: 300
---

# Lazy Install

Let's change the version of GitHub CLI and execute it.

```bash
# Change cli/cli version to v2.1.0
sed -i "s|cli/cli@.*|cli/cli@v2.1.0|" aqua.yaml
gh version
```

```console
$ gh version
INFO[0000] download and unarchive the package            aqua_version=2.16.4 env=linux/arm64 exe_name=gh package_name=cli/cli package_version=v2.1.0 program=aqua registry=standard
gh version 2.1.0 (2021-10-14)
https://github.com/cli/cli/releases/tag/v2.1.0
```

You find that `cli/cli@v2.1.0` is installed automatically.
You don't have to run `aqua i` explicitly.
We call this feature `Lazy Install`.

Note that `Lazy Install` doesn't work if the symbolic link isn't created in `${AQUA_ROOT_DIR}/bin` yet.

About Lazy Install, see also [Reference](/docs/reference/lazy-install).
