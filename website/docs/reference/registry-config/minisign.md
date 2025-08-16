---
sidebar_position: 2110
---

# minisign

- `aqua >= v2.31.0`

Please see [Reference](/docs/reference/security/minisign) too.

## Fields

- enabled (bool)
- type (string): `github_release` or `http`
- repo_owner (string) (optional):
- repo_name (string) (optional):
- asset (string) (`github_release` requires):
- url (string) (`http` requires):
- public_key (string)

e.g.

```yaml
minisign:
  type: http
  url: https://ziglang.org/builds/zig-{{.OS}}-{{.Arch}}-{{.Version}}.{{.Format}}.minisig
  public_key: "RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"
```
