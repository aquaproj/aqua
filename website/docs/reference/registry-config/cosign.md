---
sidebar_position: 2000
---

# cosign

- `aqua > v1.26.0`

Please see [Cosign and SLSA Provenance Support](/docs/reference/security/cosign-slsa) too.

## Fields

- opts ([]string): [cosign verify-blob](https://docs.sigstore.dev/signing/quickstart/#verifying-a-signed-blob) options
- signature
  - type (string): `github_release` or `http`
  - repo_owner (string) (optional):
  - repo_name (string) (optional):
  - url (string) (`http` requires):
  - asset (string) (`github_release` requires):
- key
  - same as `signature`
- certificate
  - same as `signature`
- bundle (`aqua >= v2.47.0`)
  - same as `signature`

e.g.

```yaml
cosign:
  opts:
    - --signature
    - https://github.com/terraform-linters/tflint/releases/download/{{.Version}}/checksums.txt.keyless.sig
    - --certificate
    - https://github.com/terraform-linters/tflint/releases/download/{{.Version}}/checksums.txt.pem
```

```yaml
cosign:
  signature:
    type: github_release
    asset: checksums.txt.keyless.sig
  certificate:
    type: github_release
    asset: checksums.txt.pem
```
