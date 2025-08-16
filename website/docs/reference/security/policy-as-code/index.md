---
sidebar_position: 80
---

# Policy as Code

`aqua >= v1.24.0`

[#1306](https://github.com/aquaproj/aqua/issues/1306)

See also.

- [Guides > Policy as Code](/docs/guides/policy-as-code)
- [Why is Policy needed?](why-policy-is-needed.md)
- [Git Repository root's policy file and policy commands](git-policy.md)

## Change Logs

- v2.3.0: Support `Git Repository root's policy file` and policy commands
- v2.1.0: Support `AQUA_DISABLE_POLICY`
- v2.0.0: aqua allows only Standard Registry by default
- v1.24.0: Introduce Policy

## JSON Schema

- https://github.com/aquaproj/aqua/blob/main/json-schema/policy.json
- https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/policy.json

## Disable Policy

aqua >= v2.1.0 [#1790](https://github.com/aquaproj/aqua/issues/1790)

:::caution
We don't recommend this feature basically because Policy is important in terms of security.
This feature is introduced to enable users using non Standard Registries to upgrade aqua to v2 easily.
You shouldn't use this feature in CI.
:::

If `AQUA_DISABLE_POLICY` is `true`, Policy is disabled and every Registry and Package are available.

## Policy Types

There are two types of Policies

1. [Git Repository root's policy file](git-policy.md)
1. [AQUA_POLICY_CONFIG](#aqua_policy_config)

We recommend `Git Repository root's policy file` instead of `AQUA_POLICY_CONFIG`.
`Git Repository root's policy file` was introduced to solve the issue of `AQUA_POLICY_CONFIG`.
Please see [Why is `Git Repository root's policy file` needed](git-policy.md#why-this-feature-is-needed).

## AQUA_POLICY_CONFIG

You can specify Policy file paths by the environment variable `AQUA_POLICY_CONFIG`.

e.g.

```sh
export AQUA_POLICY_CONFIG=$PWD/aqua-policy.yaml:$AQUA_POLICY_CONFIG
```

Unlike `Git Repository root's policy file`, you don't have to run `aqua policy allow` command to allow Policy files.
