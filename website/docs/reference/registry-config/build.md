---
sidebar_position: 610
---

# build

[#2132](https://github.com/aquaproj/aqua/issues/2132) [#2317](https://github.com/aquaproj/aqua/pull/2317) aqua >= [v2.13.0](https://github.com/aquaproj/aqua/releases/tag/v2.13.0)

`build` enables to install packages by `go_install` or `go_build` on platforms where prebuilt binaries aren't published.

e.g.

```yaml
packages:
  - type: github_release
    repo_owner: suzuki-shunsuke
    repo_name: tfcmt
    asset: tfcmt_{{.OS}}_{{.Arch}}.{{.Format}}
    format: tar.gz
    supported_envs:
      - linux
    build:
      type: go_build
      files:
        - name: tfcmt
          src: ./cmd/tfcmt
          dir: tfcmt-{{trimV .Version}}
```

`supported_envs` is `linux`, so on platforms other than linux aqua installs tfcmt by `go_build`.

`go_install` is also available.

```yaml
build:
  type: go_install
  path: github.com/suzuki-shunsuke/tfcmt/v4/cmd/tfcmt
```

If `go_build` failed on windows/arm64 and you'd like to exclude windows/arm64, `excluded_envs` is available.

```yaml
build:
  type: go_build
  excluded_envs:
    - windows/arm64
  files:
    - name: tfcmt
      src: ./cmd/tfcmt
      dir: tfcmt-{{trimV .Version}}
```

If you'd like to disable `build` in version_overrides, `enabled` is available.

```yaml
build:
  enabled: false
```

## Why not `overrides`?

Of course, we can do the same thing with [overrides](overrides.md).
But `build` makes the intension of the code clear and simplify the code.
