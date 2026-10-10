---
sidebar_position: 1255
---

# packslip

- `aqua >= v2.65.0`

aqua supports verifying packages with [packslip](https://github.com/jdx/packslip).
packslip publishes a signed release bundle `packslip.sigstore.json` in a release, which records the digests of the release's artifacts.

aqua downloads `packslip.sigstore.json` from the package's release and runs `packslip verify` against the downloaded asset.
aqua installs packslip automatically, so you don't need to install it.

## Example

```yaml
packages:
  - type: github_release
    repo_owner: jdx
    repo_name: hk
    # ...
    packslip:
      identity_prefix: https://github.com/jdx/hk/
      issuer: https://token.actions.githubusercontent.com
```

You can also pin the signer fingerprint or the public key.

```yaml
packslip:
  pin: ps1_snirenkjwr7m5ozgcufameodnm
```

```yaml
packslip:
  pubkey: <the base64 line of the public key>
```

You must specify at least one of `pubkey`, `identity`, `identity_prefix`, and `pin`.
Without them, packslip trusts the project the bundle claims, which doesn't show that the bundle is the package's.

Please see [the registry configuration](/docs/reference/registry-config/packslip) and [packslip's document](https://packslip.dev/docs/verifying/) too.

## aqua gr

If a release has `packslip.sigstore.json`, `aqua gr` generates the packslip setting.

## Supported platforms

packslip supports linux, darwin/arm64, and windows.
On other platforms, aqua skips verification with packslip.
