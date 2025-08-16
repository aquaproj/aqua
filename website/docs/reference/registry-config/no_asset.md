---
sidebar_position: 2140
---

# no_asset

aqua >= [v1.35.0](https://github.com/aquaproj/aqua/releases/tag/v1.35.0)

[#1693](https://github.com/aquaproj/aqua/issues/1693) [#1695](https://github.com/aquaproj/aqua/pull/1695)

:::caution
If you want to customize the error message, plse use [error_message](error_message.md).
:::

If this field is set, it fails to install the package and outputs the error message.

e.g.

registry.yaml

```yaml
packages:
  - type: github_release
    repo_owner: grafana
    repo_name: xk6
    version_constraint: semver("< 0.9.0")
    version_overrides:
      - version_constraint: semver(">= 0.9.0")
      	no_asset: true
```

```console
$ xk6 --help                   
ERRO[0000] failed to install a package grafana/xk6@v0.9.0. No asset is released in this version  aqua_version= env=darwin/arm64 exe_name=xk6 exe_path=/Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/grafana/xk6/v0.9.0/xk6_0.9.0_mac_arm64.tar.gz/xk6 package=grafana/xk6 package_name=grafana/xk6 package_version=v0.9.0 program=aqua registry=standard
FATA[0000] aqua failed                                   aqua_version= env=darwin/arm64 error= exe_name=xk6 package=grafana/xk6 package_version=v0.9.0 program=aqua
```
