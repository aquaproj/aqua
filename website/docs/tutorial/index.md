---
sidebar_position: 100
---

<!-- docfresh post
command: |
  test ! -f aqua.yaml || rm aqua.yaml
-->

# Quick Start

aqua is a CLI tool to install CLI tools with declarative YAML configuration.
In this quick start, let's install aqua and install tools with aqua.

## Demo

Please see [Demo](https://asciinema.org/a/498262?autoplay=1).

## Install aqua

[Install](install.md)

Please confirm if aqua is installed correctly.

<!-- docfresh begin
pre_command:
  command: |
    test ! -f aqua.yaml || rm aqua.yaml
command:
  command: aqua -v
-->
```sh
aqua -v
```

Output:

```
aqua version 2.56.7
```
<!-- docfresh end -->

## Docker

If you want to try this tutorial in the clean environment, container is useful.

```bash
docker run --rm -ti mirror.gcr.io/ubuntu:24.04 bash
```

<!-- docfresh container
id: foo
engine: docker-cli
image: mirror.gcr.io/ubuntu:24.04
workspace: /root/workspace
env:
  PATH: /root/.local/share/aquaproj-aqua/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
-->

<!-- docfresh begin
command:
  quiet: true
  container:
    id: foo
  command: |
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y curl vim

    mkdir ~/workspace
    cd ~/workspace
    export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
    curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer
    echo "98b883756cdd0a6807a8c7623404bfc3bc169275ad9064dc23a6e24ad398f43d  aqua-installer" | sha256sum -c -

    chmod +x aqua-installer
    ./aqua-installer
-->
```sh
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y curl vim

mkdir ~/workspace
cd ~/workspace
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer
echo "98b883756cdd0a6807a8c7623404bfc3bc169275ad9064dc23a6e24ad398f43d  aqua-installer" | sha256sum -c -

chmod +x aqua-installer
./aqua-installer
```

<!-- docfresh end -->

## Create a configuration file

Create a configuration file by `aqua init` command.

<!-- docfresh begin
command:
  command: aqua init
  quiet: true
  container:
    id: foo
-->
```sh
aqua init
```

<!-- docfresh end -->

aqua.yaml is created.

<!-- docfresh begin
code_block: true
file:
  path: aqua.yaml
  container:
    id: foo
-->
```yaml
---
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/aqua-yaml.json
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: v4.479.0  # renovate: depName=aquaproj/aqua-registry
packages:
```
<!-- docfresh end -->

`packages` is still empty, so let's add packages to install them.

:::info
See also [Split config](/docs/guides/split-config)

```sh
# -d: Create aqua/aqua.yaml
# -u: Import packages from imports/.*\.ya?ml
aqua init [-d] [-u]
```
:::

## Install tools with aqua

Let's install [GitHub Official CLI](https://cli.github.com/) and [fzf](https://github.com/junegunn/fzf) with aqua.

Add packages to `aqua.yaml`.

<!-- docfresh begin
command:
  command: aqua g -i cli/cli junegunn/fzf
  quiet: true
  container:
    id: foo
-->
```sh
aqua g -i cli/cli junegunn/fzf
```

<!-- docfresh end -->

Packages are added to the field `packages`.

<!-- docfresh begin
code_block: true
file:
  path: aqua.yaml
  range:
    start: -3
  container:
    id: foo
-->
```yaml
packages:
- name: cli/cli@v2.87.3
- name: junegunn/fzf@v0.70.0
```
<!-- docfresh end -->

Then run `aqua i`.

<!-- docfresh begin
command:
  command: aqua i
  quiet: true
  env:
    AQUA_CONIG: aqua.yaml
  container:
    id: foo
-->
```sh
aqua i
```

<!-- docfresh end -->

Congratulation! Tools are installed correctly.

<!-- docfresh begin
command:
  command: |
    command -v gh
    gh version
    command -v fzf
    fzf --version
  container:
    id: foo
-->
```sh
command -v gh
gh version
command -v fzf
fzf --version
```

Output:

```
/root/.local/share/aquaproj-aqua/bin/gh
gh version 2.87.3 (2026-02-23)
https://github.com/cli/cli/releases/tag/v2.87.3
/root/.local/share/aquaproj-aqua/bin/fzf
0.70.0 (eacef5ea)
```
<!-- docfresh end -->

aqua installs tools in `${AQUA_ROOT_DIR}`.
