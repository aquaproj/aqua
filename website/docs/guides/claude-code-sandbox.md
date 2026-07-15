---
sidebar_position: 900
---

# Use aqua with Claude Code's sandbox

[Claude Code](https://code.claude.com/docs/en/sandboxing) can run Bash commands inside an OS-level sandbox that restricts filesystem and network access.
aqua doesn't work inside the sandbox with the default settings.
This page describes the minimum settings aqua needs.

:::info
This page is about the sandbox built into Claude Code, which is enabled by `sandbox.enabled` in `settings.json`.
:::

## TL;DR

Add the following to `.claude/settings.json`:

```json
{
  "sandbox": {
    "enabled": true,
    "enableWeakerNetworkIsolation": true,
    "network": {
      "allowedDomains": [
        "github.com",
        "raw.githubusercontent.com",
        "release-assets.githubusercontent.com",
        "api.github.com"
      ]
    },
    "filesystem": {
      "allowWrite": ["~/.local/share/aquaproj-aqua"]
    }
  }
}
```

`enableWeakerNetworkIsolation` weakens the sandbox. Please read [Why enableWeakerNetworkIsolation is required](#why-enableweakernetworkisolation-is-required) before enabling it, and decide for yourself whether the trade-off is acceptable.

:::caution
Sandbox settings are read when Claude Code starts.
Editing `settings.json` in a running session has no effect. Restart Claude Code after changing them.
:::

## Requirements

aqua needs three things from the sandbox.

### 1. Write access to the root directory

By default, sandboxed commands can write only to the current working directory and the session temp directory (`$TMPDIR`).
aqua writes to its root directory (`aqua root-dir`), which is `${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua` by default, so it needs to be allowed explicitly:

```json
{
  "sandbox": {
    "filesystem": {
      "allowWrite": ["~/.local/share/aquaproj-aqua"]
    }
  }
}
```

Without this, writes fail with `Operation not permitted`.
If you change the root directory with `AQUA_ROOT_DIR`, allow that path instead.

### 2. Network access

Add the hosts aqua downloads from to `network.allowedDomains`:

| Host | Purpose |
| --- | --- |
| `github.com` | Downloads aqua-proxy and packages hosted on GitHub Releases |
| `release-assets.githubusercontent.com` | GitHub Releases assets redirect here |
| `raw.githubusercontent.com` | Standard Registry (`registry.yaml`) |
| `api.github.com` | GitHub API. Needed to resolve versions (`aqua g`, `aqua up`, unpinned versions) |

`api.github.com` isn't used when every package version is pinned and the download succeeds, but you generally want it.

Packages aren't limited to GitHub. If you install packages hosted elsewhere (`http` type packages, Go modules, npm, and so on), add those hosts too.

### 3. Access to the system TLS trust service

This is the one that isn't obvious. See below.

## Why enableWeakerNetworkIsolation is required

Allowing the hosts above isn't enough. On macOS, aqua still fails:

```
Get "https://github.com/aquaproj/aqua-proxy/releases/download/v1.2.13/aqua-proxy_darwin_arm64.tar.gz":
 tls: failed to verify certificate: x509: OSStatus -26276
```

This isn't a problem with `allowedDomains`. The host is reachable; other tools can download the same URL.

`OSStatus -26276` is [`errSecInternalComponent`](https://developer.apple.com/documentation/security/errsecinternalcomponent). On macOS, Go doesn't verify certificates itself: it calls the platform verifier in Security.framework, which talks to the system trust service (`com.apple.trustd.agent`) over Mach IPC. macOS's sandbox (Seatbelt) blocks that lookup, so verification fails before the request goes out.

This affects every Go-based CLI, not just aqua. Claude Code's documentation lists the same symptom for `gh`, `gcloud`, and `terraform` under [Troubleshooting](https://code.claude.com/docs/en/sandboxing#troubleshooting), and it's reported in [anthropics/claude-code#34876](https://github.com/anthropics/claude-code/issues/34876), which was closed as not planned.

Neither `SSL_CERT_FILE` nor `GODEBUG=x509usefallbackroots=1` helps:

- `SSL_CERT_FILE` is ignored on macOS. As [golang/go#77865](https://github.com/golang/go/issues/77865) puts it, "On darwin systems we use the platform certificate verifier, instead of the native Go one, by default". Supporting it on Darwin is an accepted proposal, so this may change in a future Go release.
- `GODEBUG=x509usefallbackroots=1` only takes effect if the binary calls [`x509.SetFallbackRoots`](https://pkg.go.dev/crypto/x509#SetFallbackRoots): "Setting x509usefallbackroots=1 without calling SetFallbackRoots has no effect". aqua doesn't embed fallback roots, so it does nothing.

The fix is to allow the sandbox to reach the trust service:

```json
{
  "sandbox": {
    "enableWeakerNetworkIsolation": true
  }
}
```

:::caution
As the name says, this weakens the sandbox. Claude Code's documentation describes it as reducing security "by opening a potential data exfiltration path": the trust service fetches OCSP/CRL data over the network without going through the sandbox's proxy, so it isn't covered by `allowedDomains`.

Decide whether this trade-off is acceptable for you. It only matters on macOS; on Linux the sandbox uses bubblewrap and this setting does nothing.
:::

`network.allowMachLookup` is equivalent:

```json
{
  "sandbox": {
    "network": {
      "allowMachLookup": ["com.apple.trustd.agent"]
    }
  }
}
```

Both grant the same Mach service lookup, and both were confirmed to work. `allowMachLookup` isn't safer just because it names the service explicitly; the trade-off is the same. Use `enableWeakerNetworkIsolation`, which is the documented setting for this purpose.

## Why excludedCommands isn't enough

Claude Code's documentation recommends `excludedCommands` for Go-based CLIs, which runs them outside the sandbox:

```json
{
  "sandbox": {
    "excludedCommands": ["aqua *"]
  }
}
```

This works for `aqua` itself, but it doesn't work well for aqua specifically, for three reasons.

aqua installs packages lazily. The first time you run an installed tool, aqua-proxy downloads it. That download runs under the tool's own command name, not `aqua`, so `aqua *` doesn't cover it and it fails with the same TLS error:

```console
$ echo 'a: 1' | yq '.a'
ERR request failed error="Get \"https://github.com/mikefarah/yq/releases/download/v4.44.3/yq_darwin_arm64\":
 tls: failed to verify certificate: x509: OSStatus -26276"
```

The pattern is matched against the command string, and the match is easy to miss. Prefixing environment variables defeats it: `AQUA_ROOT_DIR=/tmp/foo aqua i` doesn't match `aqua *`, so it runs sandboxed and fails, while `cd /tmp/foo && aqua i` does match.

When it does match, it excludes more than you might expect. The exclusion applies to the whole command, so `aqua i && rm -rf ~/.config` runs entirely outside the sandbox, including the part that has nothing to do with aqua. That's a weaker outcome than `enableWeakerNetworkIsolation`, which keeps aqua inside the sandbox and subject to `allowedDomains` and `allowWrite`.

## Notes

`allowedDomains` isn't necessarily a hard boundary. By default, when a sandboxed command reaches a host that isn't allowed, Claude Code prompts you, and approving it allows that host for the rest of the session. If you use [auto mode](https://code.claude.com/docs/en/permission-modes), the classifier may approve hosts without prompting, so hosts outside `allowedDomains` can still be reached. To make it a hard allowlist, an administrator sets `network.allowManagedDomainsOnly` in managed settings.
