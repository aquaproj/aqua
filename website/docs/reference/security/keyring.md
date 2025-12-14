---
sidebar_position: 410
---

# Manage a GitHub access token using Keyring

:::info
We recommend [ghtkn integration](./ghtkn.md) because it is securer than managing a long-lived personal access token using Keyring.
:::

aqua >= v2.51.0 [#3852](https://github.com/aquaproj/aqua/pull/3852)

You can manage a GitHub Access token using secret store such as [Windows Credential Manager](https://support.microsoft.com/en-us/windows/accessing-credential-manager-1b5c916a-6a16-889f-8581-fc16e8165ac0), [macOS Keychain](https://en.wikipedia.org/wiki/Keychain_(software)), and [GNOME Keyring](https://wiki.gnome.org/Projects/GnomeKeyring).

1. Configure a GitHub Access token by `aqua token set` command:

```console
$ aqua token set
Enter a GitHub access token: # Input GitHub Access token
```

or you can also pass a GitHub Access token via standard input:

```sh
echo "<github access token>" | aqua tokn set -stdin
```

2. Enable the feature by setting the environment variable `AQUA_KEYRING_ENABLED`:

```sh
export AQUA_KEYRING_ENABLED=true
```

Note that if the environment variable `GITHUB_TOKEN` is set, this feature gets disabled.

You can remove a GitHub Access token from keyring by `aqua token rm` command:

```sh
aqua token rm
```

:::info
**GitHub Enterprise Server**: Keyring is not supported for GHES. Use environment variables instead. See [GHES Guide](/docs/guides/github-enterprise-server) for details.
:::
