---
sidebar_position: 405
---

# ghtkn integration

aqua >= v2.54.0 [#4173](https://github.com/aquaproj/aqua/pull/4173)

aqua can get or create GitHub App User Access Tokens to call GitHub APIs by [ghtkn](https://github.com/suzuki-shunsuke/ghtkn).
For the local development, you don't need to use personal access tokens anymore.

## Limitations

This feature creates access tokens via [GitHub App's Device Flow](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app#using-the-device-flow-to-generate-a-user-access-token).
So this doesn't work on non-interactive environment such as CI.

## Requirements

- Keyring such as [Windows Credential Manager](https://support.microsoft.com/en-us/windows/accessing-credential-manager-1b5c916a-6a16-889f-8581-fc16e8165ac0), [macOS Keychain](https://en.wikipedia.org/wiki/Keychain_(software)), and [GNOME Keyring](https://wiki.gnome.org/Projects/GnomeKeyring)
- GitHub Apps and their client ids

Note that ghtkn CLI isn't required.

## How To Set up

1. Create ghtkn's configuration file

[For more details, please see the document of ghtkn.](https://github.com/suzuki-shunsuke/ghtkn).

2. Update aqua to v2.54.0 or later
3. Set the environment variable

```sh
export AQUA_GHTKN_ENABLED=true
```

Remove environment variables such as `GITHUB_TOKEN`, `AQUA_GITHUB_TOKEN`, and `AQUA_KEYRING_ENABLED`.
