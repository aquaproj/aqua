---
sidebar_position: 300
---

# update-checksum-workflow

https://github.com/aquaproj/update-checksum-workflow

GitHub Actions Reusable Workflow to update aqua-checksums.json
If aqua-checksums.json isn't latest, update aqua-checksums.json and push a commit

About aqua's Checksum Verification, please see [the document](/docs/reference/security/checksum) too.

## :warning: Please consider autofix.ci or Securefix Action

[We recommend autofix.ci or Securefix Action rather than this workflow.](/docs/guides/checksum#recommend-autofixci-or-securefix-action-instead-of-update-checksum-action-and-update-checksum-workflow)

## Workflow

[Workflow](https://github.com/aquaproj/update-checksum-workflow/blob/main/.github/workflows/update-checksum.yaml)

## Requirements

Nothing.

:::info
As of update-checksum-workflow v1.0.3, [ghcp](https://github.com/int128/ghcp) isn't necessary.
:::

### Example

```yaml
name: update-aqua-checksum
on:
  pull_request:
    paths:
      - aqua.yaml
      - aqua-checksums.json
jobs:
  update-aqua-checksums:
    uses: aquaproj/update-checksum-workflow/.github/workflows/update-checksum.yaml@d248abb88efce715d50eb324100d9b29a20f7d18 # v1.0.4
    permissions:
      contents: read
    with:
      aqua_policy_config: aqua-policy.yaml
      aqua_version: v2.48.3
      prune: true
    secrets:
      gh_app_id: ${{secrets.APP_ID}}
      gh_app_private_key: ${{secrets.APP_PRIVATE_KEY}}
```
