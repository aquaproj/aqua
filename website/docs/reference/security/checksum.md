---
sidebar_position: 50
---

# Checksum Verification

`aqua >= v1.20.0`

[#427](https://github.com/aquaproj/aqua/issues/427)

Checksum Verification is a feature verifying downloaded assets with checksum.
Checksum Verification prevents the supply chain attack and allows you to install tools securely.

## See also

- [Tutorial](/docs/guides/checksum)
- [Configuration](/docs/reference/config/checksum)
- [Registry Configuration](/docs/reference/registry-config/checksum)
- Blogs
  - [2022-11-08 Checksum Verification by aqua](https://dev.to/suzukishunsuke/checksum-verification-by-aqua-5038)
  - [2022-10-10 aqua CLI Version Manager が checksum の検証をサポート](https://zenn.dev/shunsuke_suzuki/articles/aqua-checksum-verification)

## How it works

When a tool is installed, aqua verifies the checksum as the following.

1. Download the tool in the temporal directory
1. Calculate the checksum from the downloaded file
1. Get the expected checksum
1. If the actual checksum is different from the expected checksum, make the installation failure
1. If the checksum isn't found in `aqua-checksums.json`, the expected checksum is added to `aqua-checksums.json`
1. Install the tool

aqua gets the expected checksum from the following sources.

1. `aqua-checksums.json`
1. checksum files that each tools publish
1. If the tool doesn't publish checksum files, aqua treats the checksum calculated from the downloaded asset as the expected checksum

e.g. `aqua-checksums.json`

```json
{
  "checksums": [
    {
      "id": "github_release/github.com/golangci/golangci-lint/v1.49.0/golangci-lint-1.49.0-darwin-amd64.tar.gz",
      "checksum": "20cd1215e0420db8cfa94a6cd3c9d325f7b39c07f2415a02d111568d8bc9e271",
      "algorithm": "sha256"
    },
    {
      "id": "github_release/github.com/golangci/golangci-lint/v1.49.0/golangci-lint-1.49.0-darwin-arm64.tar.gz",
      "checksum": "cabb1a4c35fe1dadbe5a81550a00871281a331e7660cd85ae16e936a7f0f6cfc",
      "algorithm": "sha256"
    }
  ]
}
```

Many tools publish checksum files, so aqua gets checksums from them.

e.g.

* [Terraform](https://releases.hashicorp.com/terraform/1.2.7/terraform_1.2.7_SHA256SUMS)
* [GitHub CLI](https://github.com/cli/cli/releases/download/v2.14.4/gh_2.14.4_checksums.txt)

If no checksum file for a tool is published, aqua can also get checksums by downloading assets and calculating checksums.


## aqua-registry version

From [v3.90.0](https://github.com/aquaproj/aqua-registry/releases/tag/v3.90.0), aqua-registry supports the checksum verification.

## Remove unused checksums with `-prune` option

aqua >= [v1.28.0](https://github.com/aquaproj/aqua/releases/tag/v1.28.0)

When tools are updated, checksums for old versions are basically unneeded.
Or when we remove some tools from `aqua.yaml`, checksums for those tools would be unneeded.

You can remove unused checksums by setting `-prune` option.

```
aqua update-checksum -prune
```

## Verify checksums of Registries

aqua >= [v1.30.0](https://github.com/aquaproj/aqua/releases/tag/v1.30.0)

[#1491](https://github.com/aquaproj/aqua/issues/1491) [#1508](https://github.com/aquaproj/aqua/pull/1508)

aqua verifies checksums of Registries if Checksum Verification is enabled.

aqua.yaml

```yaml
checksum:
  enabled: true
```

aqua-checksums.json

```json
{
  "checksums": [
    {
      "id": "registries/github_content/github.com/aquaproj/aqua-registry/v3.114.0/registry.yaml",
      "checksum": "b5b922c4d64609e536daffec6e480d0fed3ee56b16320a10c38ae12df7f045e8b20a0c05ec66eb28146cee42559e5e6c4e4bc49ce89ffe48a5640999cc6248bd",
      "algorithm": "sha512"
    }
  ]
}
```

If the checksum is invalid, it would fail to install Registries.

```
ERRO[0000] install the registry                          actual_checksum=b5b922c4d64609e536daffec6e480d0fed3ee56b16320a10c38ae12df7f045e8b20a0c05ec66eb28146cee42559e5e6c4e4bc49ce89ffe48a5640999cc6248be aqua_version= env=darwin/arm64 error="check a registry's checksum: checksum is invalid" exe_name=starship expected_checksum=b5b922c4d64609e536daffec6e480d0fed3ee56b16320a10c38ae12df7f045e8b20a0c05ec66eb28146cee42559e5e6c4e4bc49ce89ffe48a5640999cc6248bd program=aqua registry_name=standard
FATA[0000] aqua failed                                   aqua_version= env=darwin/arm64 error="it failed to install some registries" exe_name=starship program=aqua
```

## Generate checksum configuration automatically

It is bothersome to write the checksum configuration manually, so aqua supports scaffolding the configuration.

[aqua gr](/docs/develop-registry#scaffold-registry-configuration) scaffolds the checksum configuration too.

:::caution
The scaffolding isn't perfect, so sometimes you have to fix the code manually.
:::

## Enforce Checksum Verification by environment variables

aqua >= v2.27.0 [#2702](https://github.com/aquaproj/aqua/issues/2702) [#2806](https://github.com/aquaproj/aqua/pull/2806)

You can enforce checksum verification by environment variables `AQUA_ENFORCE_CHECKSUM` and `AQUA_ENFORCE_REQUIRE_CHECKSUM`.

```sh
export AQUA_ENFORCE_CHECKSUM=true
export AQUA_ENFORCE_REQUIRE_CHECKSUM=true
```

This is useful for both CI and local development.

Checksum verification is disabled by default, and you can disable checksum verification by setting.
If you manage a Monorepo and want to make checksum verification mandatory in CI, you can set these environment variables in CI. Then checksum verification is enabled regardless of the setting of aqua.yaml.

And if you want to enforce checksum verification on your laptop, you can set these environment variables in your shell configuration files such as .bashrc and .zshrc.

## Question: Should `aqua-checksums.json` be managed with Git?

Yes. You should manage `aqua-checksums.json` with Git.
