---
sidebar_position: 400
---

<!-- docfresh container
id: foo
engine: docker-cli
image: mirror.gcr.io/ubuntu:24.04
workspace: /root/workspace
env:
  PATH: /root/.local/share/aquaproj-aqua/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
command:
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
    aqua init
    aqua g -i cli/cli junegunn/fzf
    aqua i
    aqua g -i tfmigrator/cli
-->

# `aqua i`'s `-l` option

You added [tfmigrator/cli](https://github.com/tfmigrator/cli) in [Search packages](search-packages.md), but it isn't installed yet.

<!-- docfresh begin
command:
  ignore_fail: true
  hide_output: true
  container:
    id: foo
  command: |
    command -v tfmigrator
-->
```sh
command -v tfmigrator
```

<!-- docfresh end -->

Let's run `aqua i -l`.

<!-- docfresh begin
command:
  container:
    id: foo
  command: |
    aqua i -l
-->
```sh
aqua i -l
```

Output:

```
Mar 10 00:26:11.009 INF create a symbolic link program=aqua version=2.56.7 package_name=tfmigrator/cli package_version=v0.2.2 command=tfmigrator
```
<!-- docfresh end -->

The command would exit immediately because the tool isn't downloaded and installed yet.

The command `aqua i` installs all tools at once.
But when the option `-l` is set, `aqua i` creates only symbolic links in `${AQUA_ROOT_DIR}/bin` and skips downloading and installing tools.

Even if downloading and installing are skipped, you can execute the tool thanks for [Lazy Install](lazy-install.md).

<!-- docfresh begin
command:
  container:
    id: foo
  command: |
    tfmigrator -v
transform:
  CombinedOutput: '{{ regexReplaceAll "[A-Z][a-z]{2}\\s+\\d{1,2}\\s\\d{2}:\\d{2}:\\d{2}\\.\\d{3}" .CombinedOutput "Jan  2 15:04:05.999" }}'
-->
```sh
tfmigrator -v
```

Output:

```
Jan  2 15:04:05.999 INF download and unarchive the package program=aqua version=2.56.7 exe_name=tfmigrator package_name=tfmigrator/cli package_version=v0.2.2 registry=standard
Jan  2 15:04:05.999 INF verify a package with slsa-verifier program=aqua version=2.56.7 exe_name=tfmigrator package_name=tfmigrator/cli package_version=v0.2.2 registry=standard
Jan  2 15:04:05.999 INF download and unarchive the package program=aqua version=2.56.7 exe_name=tfmigrator package_name=tfmigrator/cli package_version=v0.2.2 registry=standard package_name=slsa-framework/slsa-verifier package_version=v2.7.1 registry=""
Verified signature against tlog entry index 11189344 at URL: https://rekor.sigstore.dev/api/v1/log/entries/24296fb24b8ad77ab59a542e7568d3a61c0d9416c9e14f1bf4e6bfa9976e6642a33216aa9765c80a
Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.4.0" at commit 15664e1d52bf92dddaf10e4f393ed3591a9cb891
Verifying artifact /tmp/652128706: PASSED

PASSED: SLSA verification passed
tfmigrator version 0.2.2 (15664e1d52bf92dddaf10e4f393ed3591a9cb891)
```
<!-- docfresh end -->

`-l` option is useful for local development because you can install only tools which are needed for you.
