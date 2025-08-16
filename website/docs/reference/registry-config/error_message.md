---
sidebar_position: 2150
---

# error_message

aqua >= [v1.35.0](https://github.com/aquaproj/aqua/releases/tag/v1.35.0)

[#1684](https://github.com/aquaproj/aqua/issues/1684) [#1687](https://github.com/aquaproj/aqua/pull/1687)

:::caution
Please consider using [no_asset](no_asset.md) if you don't have to customize the error message.
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
        error_message: |
          From version xk6 v0.9.0 there are no more binaries published.

          https://github.com/grafana/xk6/issues/60
```

```console
$ xk6 --help                   
ERRO[0000] failed to install a package grafana/xk6@v0.9.0. From version xk6 v0.9.0 there are no more binaries published.

https://github.com/grafana/xk6/issues/60  aqua_version= env=darwin/arm64 exe_name=xk6 exe_path=/Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/github_release/github.com/grafana/xk6/v0.9.0/xk6_0.9.0_mac_arm64.tar.gz/xk6 package=grafana/xk6 package_name=grafana/xk6 package_version=v0.9.0 program=aqua registry=standard
FATA[0000] aqua failed                                   aqua_version= env=darwin/arm64 error= exe_name=xk6 package=grafana/xk6 package_version=v0.9.0 program=aqua
```
