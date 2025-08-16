---
sidebar_position: 250
---

# update-checksum-action

https://github.com/aquaproj/update-checksum-action

GitHub Actions to update aqua-checksums.json. If aqua-checksums.json isn't latest, update aqua-checksums.json and push a commit

About aqua's Checksum Verification, please see [the document](/docs/reference/security/checksum) too.

## :warning: Please consider autofix.ci or Securefix Action

[We recommend autofix.ci or Securefix Action rather than this action.](/docs/guides/checksum#recommend-autofixci-or-securefix-action-instead-of-update-checksum-action-and-update-checksum-workflow)

## Reusable Workflow

Please see [update-checksum-workflow](https://github.com/aquaproj/update-checksum-workflow).

## Requirements

[aqua](https://aquaproj.github.io/)

:::info
As of update-checksum-action v0.2.5, [ghcp](https://github.com/int128/ghcp) isn't necessary.
:::

## Example

[Workflow](https://github.com/aquaproj/example-update-checksum/blob/main/.github/workflows/test.yaml)

## Inputs

- `working_directory`: The working directory where `aqua update-checksum` is executed. If this input is not specified, the command is run on the current working directory
- `prune`: If this input is `true`, `aqua update-checksum` is executed with `-prune` option. This option removes unused checksums from the checksum file. If this input is not specified, `false` is used.
- `skip_push`: If this input is `true`, the action checks if the checksum file is up-to-date, but does not push a commit to update it. If this input is not specified, `false` is used.
- `read_checksum_token`: This token overrides `AQUA_GITHUB_TOKEN` to executes `aqua update-checksum`. It must have `contents:read` permission about all repositories in tools managed by `aqua`. This input is useful to fetch checksum from private repositories.
- `securefix_action_server_repository`: The GitHub repository for the [Securefix Action server](https://github.com/csm-actions/securefix-action). If this is set, this action uses Securefix Action to update aqua-checksums.json.
- `securefix_action_app_id`: The GitHub App ID for the Securefix Action client.
- `securefix_action_app_private_key`: The GitHub App private key for the Securefix Action client.

## Required Environment Variables

- `GITHUB_TOKEN`: GitHub Access Token. This is used to push a commit.

Required permissions: `contents: write`

:::info
~~Unfortunately, fine-grained personal access token is unavailable at the moment because it doesn't support GraphQL API.~~
~~https://github.com/cli/cli/issues/6680~~

2023-04-27 [fine-grained access token supports GraphQL API now.](https://github.blog/changelog/2023-04-27-graphql-improvements-for-fine-grained-pats-and-github-apps/)
:::

## Outputs

Nothing.
