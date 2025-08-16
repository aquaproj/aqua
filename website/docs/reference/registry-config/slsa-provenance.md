---
sidebar_position: 2100
---

# slsa_provenance

- `aqua > v1.26.0`

Please see [Cosign and SLSA Provenance Support](/docs/reference/security/cosign-slsa) too.

## Fields

- type (string): `github_release` or `http`
- repo_owner (string) (optional):
- repo_name (string) (optional):
- url (string) (`http` requires):
- asset (string) (`github_release` requires):

e.g.

```yaml
slsa_provenance:
  type: github_release
  asset: multiple.intoto.jsonl
```
