---
sidebar_position: 300
---

# Settings File

aqua >= v2.64.0 [#5279](https://github.com/aquaproj/aqua/pull/5279)

User level settings such as the log level or whether checksum verification is enforced
can be written in a settings file instead of being set as environment variables every
time.

The file is optional. Without one, every setting keeps its default.

:::info
The settings file is not `aqua.yaml`. `aqua.yaml` describes the packages of a project
and is committed to its repository; the settings file describes how you want aqua to
behave and belongs to you rather than to a project.
:::

## Path

* Linux and macOS: `${XDG_CONFIG_HOME:-$HOME/.config}/aquaproj-aqua/config.yaml`
* Windows: `aquaproj-aqua\config.yaml` in the configuration directory, which is
  `%XDG_CONFIG_HOME%` when it is set and `%LOCALAPPDATA%` otherwise

The path has no environment variable of its own.

## Example

```yaml
checksum:
  enabled: true
  require: true
  enforce: true
  enforce_require: true
log:
  level: debug
  color: always
policy:
  enabled: true
  config: /home/foo/policy.yaml
lazy_install: true
tracking: true
progress_bar: true
max_parallelism: 10
global_config: /home/foo/aqua-global.yaml
root_dir: /home/foo/.local/share/aquaproj-aqua
```

JSON Schema: https://github.com/aquaproj/aqua/blob/main/json-schema/config.json

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/aquaproj/aqua/main/json-schema/config.json
```

## Precedence

A setting can come from three places, and they beat each other in this order:

1. a command line flag
1. an environment variable
1. the settings file

So a setting in the file is the default you have chosen, and an environment variable
still overrides it for one command.

## Settings

The keys are positive. Where the environment variable is named `AQUA_DISABLE_*`, the
key is the opposite of it: `lazy_install: false` is `AQUA_DISABLE_LAZY_INSTALL=true`.

| settings file | environment variable |
| --- | --- |
| `checksum.enabled` | `AQUA_CHECKSUM` |
| `checksum.require` | `AQUA_REQUIRE_CHECKSUM` |
| `checksum.enforce` | `AQUA_ENFORCE_CHECKSUM` |
| `checksum.enforce_require` | `AQUA_ENFORCE_REQUIRE_CHECKSUM` |
| `log.level` | `AQUA_LOG_LEVEL` |
| `log.color` | [`AQUA_LOG_COLOR`](log-color.md) |
| `policy.enabled` | `AQUA_DISABLE_POLICY` (opposite) |
| `policy.config` | `AQUA_POLICY_CONFIG` |
| `lazy_install` | `AQUA_DISABLE_LAZY_INSTALL` (opposite) |
| `tracking` | `AQUA_DISABLE_TRACKING` (opposite) |
| `progress_bar` | [`AQUA_PROGRESS_BAR`](progress-bar.md) |
| `max_parallelism` | `AQUA_MAX_PARALLELISM` |
| `global_config` | [`AQUA_GLOBAL_CONFIG`](/docs/tutorial/global-config) |
| `root_dir` | `AQUA_ROOT_DIR` |

`checksum` is described in [Checksum Verification](checksum.md), and `policy` in
[Policy as Code](/docs/reference/security/policy-as-code).

Some environment variables have no key on purpose:

* `AQUA_GITHUB_TOKEN` and `GITHUB_TOKEN`, because a settings file is a plain text file.
  A token belongs in [Keyring](/docs/reference/security/keyring) or
  [ghtkn](/docs/reference/security/ghtkn) instead
* `AQUA_GOOS`, `AQUA_GOARCH` and `AQUA_LIBC`, which pretend to be another platform for
  one command. Written down in a file they would apply to every command
* `AQUA_X_SYS_EXEC` and `AQUA_EXPERIMENTAL_X_SYS_EXEC`

## An unknown key is ignored

A settings file is written once and then carried from machine to machine, so it outlives
the version of aqua that wrote it. Refusing a key that a newer aqua understands would
break every older aqua the file reaches, so unknown keys are ignored rather than
rejected.

A misspelled key is therefore ignored too. The JSON Schema above catches that in an
editor.

A file that isn't valid YAML is an error. Ignoring it would drop a setting such as
`checksum.enforce`, and you would run without the verification you asked for and never
hear about it.
