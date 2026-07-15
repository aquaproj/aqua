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

On Linux and WSL2, add the following to `.claude/settings.json`:

```json
{
  "sandbox": {
    "enabled": true,
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

On macOS, add `env.GODEBUG` as well:

```json
{
  "env": {
    "GODEBUG": "x509usefallbackroots=1"
  },
  "sandbox": {
    "...": "same as above"
  }
}
```

`GODEBUG` works around a TLS verification failure that only happens on macOS, so it isn't needed on Linux or WSL2. Don't set it there: it would replace the system's certificate store with the roots embedded in aqua for no benefit.

It requires aqua [v2.62.0](https://github.com/aquaproj/aqua/releases/tag/v2.62.0) or later [#5024](https://github.com/aquaproj/aqua/pull/5024). For older versions, see [Older versions of aqua](#older-versions-of-aqua).

Don't set it either if you rely on a custom CA in the system trust store, such as behind a TLS-inspecting proxy. See [the caution below](#why-godebug-is-needed).

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

### 3. A way to verify TLS certificates (macOS only)

This is the one that isn't obvious. See below.
On Linux and WSL2, Go verifies certificates itself and the two requirements above are enough.

## Why GODEBUG is needed

Allowing the hosts above isn't enough. On macOS, aqua still fails:

```
Get "https://github.com/aquaproj/aqua-proxy/releases/download/v1.2.13/aqua-proxy_darwin_arm64.tar.gz":
 tls: failed to verify certificate: x509: OSStatus -26276
```

This isn't a problem with `allowedDomains`. The host is reachable; other tools can download the same URL.

`OSStatus -26276` is [`errSecInternalComponent`](https://developer.apple.com/documentation/security/errsecinternalcomponent). On macOS, Go doesn't verify certificates itself: it calls the platform verifier in Security.framework, which talks to the system trust service (`com.apple.trustd.agent`) over Mach IPC. macOS's sandbox (Seatbelt) blocks that lookup, so verification fails before the request goes out.

This affects every Go-based CLI, not just aqua. Claude Code's documentation lists the same symptom for `gh`, `gcloud`, and `terraform` under [Troubleshooting](https://code.claude.com/docs/en/sandboxing#troubleshooting), and it's reported in [anthropics/claude-code#34876](https://github.com/anthropics/claude-code/issues/34876), which was closed as not planned.

Since [v2.62.0](https://github.com/aquaproj/aqua/releases/tag/v2.62.0), aqua embeds a copy of the Mozilla root certificates [#5024](https://github.com/aquaproj/aqua/pull/5024). Setting `GODEBUG=x509usefallbackroots=1` makes Go verify certificates with its own pure Go verifier and those embedded roots, so it never talks to the trust service and the sandbox doesn't need to be weakened:

```json
{
  "env": {
    "GODEBUG": "x509usefallbackroots=1"
  }
}
```

Setting it in `env` applies it to every Bash command in the session, which is what you want: aqua also downloads packages when you run an installed tool for the first time, and that runs under the tool's own name rather than `aqua`.

:::caution
`GODEBUG=x509usefallbackroots=1` makes Go ignore the system trust store, on every platform, and trust only the roots embedded in the binary. If you are behind a TLS-inspecting proxy, or otherwise rely on a CA added to the system trust store, aqua fails to verify certificates with this set. On macOS, use [`enableWeakerNetworkIsolation`](#older-versions-of-aqua) instead. This is also why you shouldn't set it on Linux or WSL2, where it buys you nothing.

Setting it session-wide is harmless for other Go tools. `x509usefallbackroots=1` has no effect unless the binary embeds fallback roots, which most tools don't.
:::

`SSL_CERT_FILE` is not an alternative: it's ignored on macOS. As [golang/go#77865](https://github.com/golang/go/issues/77865) puts it, "On darwin systems we use the platform certificate verifier, instead of the native Go one, by default". Supporting it on Darwin is an accepted proposal, so this may change in a future Go release.

## Older versions of aqua

Before v2.62.0, aqua doesn't embed fallback roots, and `GODEBUG=x509usefallbackroots=1` does nothing: per [`x509.SetFallbackRoots`](https://pkg.go.dev/crypto/x509#SetFallbackRoots), "Setting x509usefallbackroots=1 without calling SetFallbackRoots has no effect".

The only remaining option is to let the sandbox reach the trust service:

```json
{
  "sandbox": {
    "enableWeakerNetworkIsolation": true
  }
}
```

This is also the option to use if you can't set `GODEBUG` because you rely on a custom CA.

:::caution
As the name says, this weakens the sandbox. Claude Code's documentation describes it as reducing security "by opening a potential data exfiltration path": the trust service fetches OCSP/CRL data over the network without going through the sandbox's proxy, so it isn't covered by `allowedDomains`.

It only matters on macOS. On Linux the sandbox uses bubblewrap, and this setting does nothing.
:::

`network.allowMachLookup: ["com.apple.trustd.agent"]` is equivalent. Both grant the same Mach service lookup, and `allowMachLookup` isn't safer just because it names the service explicitly; the trade-off is the same.

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

When it does match, it excludes more than you might expect. The exclusion applies to the whole command, so `aqua i && rm -rf ~/.config` runs entirely outside the sandbox, including the part that has nothing to do with aqua.

## Notes

`allowedDomains` isn't necessarily a hard boundary. By default, when a sandboxed command reaches a host that isn't allowed, Claude Code prompts you, and approving it allows that host for the rest of the session. If you use [auto mode](https://code.claude.com/docs/en/permission-modes), the classifier may approve hosts without prompting, so hosts outside `allowedDomains` can still be reached. To make it a hard allowlist, an administrator sets `network.allowManagedDomainsOnly` in managed settings.
