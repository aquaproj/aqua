---
sidebar_position: 100
---

# Git Repository root's policy file and policy commands

`aqua >= v2.3.0`

[#1789](https://github.com/aquaproj/aqua/issues/1789) [1808](https://github.com/aquaproj/aqua/pull/1808)

`Git Repository root's policy file` is a Policy file in the Git Repository root directory.

`Git Repository root's policy file` must be located in one of the following paths from the Git Repository root directory.

- aqua-policy.yaml
- .aqua-policy.yaml
- aqua/aqua-policy.yaml
- .aqua/aqua-policy.yaml

:::caution
The file extension `.yml` isn't supported at the moment.
:::

Before aqua executes or installs packages, aqua searches `Git Repository root's policy file`.
aqua searches the Git repository root directory from the current directory to the root directory.

- If `Git Repository root's policy file` isn't found, it is same as usual.
- If `Git Repository root's policy file` is found, aqua checks if the policy file is already allowed or not.
- If `Git Repository root's policy file` is already allowed, aqua uses `Git Repository root's policy file` as Policy.
- If `Git Repository root's policy file` isn't allowed, aqua outputs the warning and ignores `Git Repository root's policy file`.

`aqua policy allow` command is a command to allow a policy file.

```console
$ aqua policy allow [<policy file path>]
```

If no argument is given, aqua allows `Git Repository root's policy file`.

Even if you allow a policy file once, you have to allow the policy file again if the policy file is modified.

:::caution
Before you run `aqua policy allow` command, you should confirm the content of aqua-policy.yaml.
If untrusted Registries are used, you shouldn't run `aqua policy allow`.
:::

`aqua policy deny` command is a command to deny a policy file.

```console
$ aqua policy deny [<policy file path>]
```

If no argument is given, aqua allows `Git Repository root's policy file`.

`aqua policy deny` is used to ignore `Git Repository root's policy file` and suppress the warning.

:::info
aqua searches `Git Repository root's policy file` per `aqua.yaml`. aqua searches `Git Repository` based on the directory where `aqua.yaml` is located.
:::

## How to use

1. Add `Git Repository root's policy file` to your Git repository
1. Run `aqua policy allow` in the repository

Please see [Getting Started](/docs/guides/policy-as-code).

## Why this feature is needed

To improve the user experience of non Standard Registries.
To set up Policy easily keeping the security.

To use non Standard Registries, you had to set the environment variable `AQUA_POLICY_CONFIG`.
But it is bothersome, especially in the team development because all members have to set the environment variable `AQUA_POLICY_CONFIG`.
Some tools such as `direnv` are useful to set environment variables, but it is undesirable to ask users to install additional tools for aqua.

So we would like to apply a policy without `AQUA_POLICY_CONFIG`, but at the same time we have to keep the security.

## Design consideration

Sometimes security and convenience are conflicted, so we have to be careful not to harm security for convenience.
To keep the security, I think aqua should ask users to allow `Git Repository root's policy file` explicitly.
This means aqua should not apply `Git Repository root's policy file` without user's approval.
So aqua asks users to allow `Git Repository root's policy file` using `aqua policy allow` command.

:::info
Unlike `Git Repository root's policy file`, aqua uses policy files in `AQUA_POLICY_CONFIG` without your approval.
Because

- To keep the compatibility
- Unlike `Git Repository root's policy file`, the environment variable `AQUA_POLICY_CONFIG` is set by you, so aqua regards `AQUA_POLICY_CONFIG` as your approval
:::
