---
sidebar_position: 1260
---

# GitHub Immutable Releases

- `aqua >= v2.55.0` [#4195](https://github.com/aquaproj/aqua/pull/4195)

[You can verify packages' GitHub Release Attestations if they are provided.](https://docs.github.com/en/code-security/supply-chain-security/understanding-your-software-supply-chain/verifying-the-integrity-of-a-release)

## How to verify packages

You don't have to do any special things.
If package versions are immutable and registries are configured, packages are verified when you install them.
aqua uses [GitHub CLI](https://cli.github.com/) internally, but aqua installs it in `$(aqua root-dir)` automatically, so you don't have to install it.

### Disable the verification of GitHub Artifact Attestations

We recommend enabling the verification for security, but you can disable the verification by the environment variable.

```sh
export AQUA_DISABLE_GITHUB_IMMUTABLE_RELEASE=true
```

## Registry Settings

e.g.

```yaml
packages:
  - type: github_release
    repo_owner: suzuki-shunsuke
    repo_name: tfcmt
    asset: tfcmt_{{.OS}}_{{.Arch}}.{{.Format}}
    format: tar.gz
    github_immutable_release: true
```
