---
sidebar_position: 500
---

# Use aqua combined with other version manager such as asdf

If you use aqua combined with other version manager such as asdf,
you should add `${AQUA_ROOT_DIR}/bin` to the environment variable `PATH` after other version manager.

e.g.

:thumbsup:

```bash
. $HOME/.asdf/asdf.sh

export PATH=${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH
```

:thumbsdown:

```bash
export PATH=${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH

. $HOME/.asdf/asdf.sh
```

This isn't a problem of aqua.
aqua is designed to allow to use aqua combined with other version managers, but many other version manager aren't.

Please assume the following usecase.
You develop the project A and B.
In the project A [Waypoint](https://www.waypointproject.io/) is managed with asdf, and in the project B Waypoint is managed with aqua.

```
project-a/
  .tool-versions # Manage Waypoint with asdf
project-b/
  aqua.yaml # Manage Waypoint with aqua
```

project-a/.tool-versions

```
waypoint v0.6.3
```

project-b/aqua.yaml

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry

packages:
- name: hashicorp/waypoint@v0.6.2
```

If you configure .bash_profile as the following,
you can manage Waypoint with asdf in the project A, but you can't manage Waypoint with aqua in the project B.

```bash
export PATH=${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH

. $HOME/.asdf/asdf.sh
```

```console
$ cd project-b
$ waypoint --version
No version is set for command waypoint
Consider adding one of the following versions in your config file at 
waypoint 0.6.3
```

This is because asdf is used in the project-b too.

On the other hand, if you configure .bash_profile as the following,
you can manage Waypoint with asdf in the project A, and manage Waypoint with aqua in the project B.

```bash
. $HOME/.asdf/asdf.sh

export PATH=${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH
```

```console
$ cd project-a
$ waypoint --version
CLI: v0.6.3 (bd303e12)

$ cd ../project-b
$ waypoint --version
CLI: v0.6.2 (99350730)
```

This is because if aqua can't find the command in the configuration files aqua finds the command from the environment variable `PATH`.
