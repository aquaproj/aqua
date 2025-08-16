---
sidebar_position: 1950
---

# supported_envs

[#882](https://github.com/aquaproj/aqua/issues/882) [#884](https://github.com/aquaproj/aqua/pull/884)

aqua >= v1.12.0

You can specify the list of supported environments (GOOS and GOARCH) in `supported_envs`.

`supported_envs` has been introduced to solve the following `supported_if` problems.

1. Complicated. There are multiple ways to express the same environments. The expression is too flexible 
1. Performance. aqua has to compile and evaluate the expression per tool. It may affect the performance although we should do the performance test

## rosetta2 and windows_arm_emulation

If [rosetta2](rosetta2.md) is `true`, `darwin/amd64` and `darwin/arm64` are supported.
If [windows_arm_emulation](windows_arm_emulation.md) is `true`, `windows/amd64` and `windows/arm64` are supported.

## Example Code

AS IS

```yaml
packages:
  - type: github_release
    repo_owner: gruntwork-io
    repo_name: terragrunt
    asset: terragrunt_{{.OS}}_{{.Arch}}
    supported_if: not (GOOS == "windows" and GOARCH == "arm64")
```

TO BE

```yaml
packages:
  - type: github_release
    repo_owner: gruntwork-io
    repo_name: terragrunt
    asset: terragrunt_{{.OS}}_{{.Arch}}
    supported_envs:
      - windows/amd64
      - darwin
      - linux
```

The following patterns are supported.

* `<GOOS>`
* `<GOOS>/<GOARCH>`

```yaml
supported_envs: [] # no environment is supported
```

```yaml
supported_envs: ["all"] # all environments are supported
```
