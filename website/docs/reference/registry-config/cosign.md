---
sidebar_position: 2000
---

# cosign

- `aqua > v1.26.0`

Please see [Cosign and SLSA Provenance Support](/docs/reference/security/cosign-slsa) too.

## Fields

- opts ([]string): [cosign verify-blob](https://docs.sigstore.dev/signing/quickstart/#verifying-a-signed-blob) options
- signature
  - type (string): `github_release`, `http`, `forgejo_release`, `gitea_release` or `gitlab_release`
  - repo_owner (string) (optional):
  - repo_name (string) (optional):
  - url (string) (`http` requires):
  - asset (string) (every type but `http` requires):
- key
  - same as `signature`
- certificate
  - same as `signature`
- bundle (`aqua >= v2.47.0`)
  - same as `signature`
- certificate_identity (string): `aqua >= v2.65.0`
- certificate_identity_regexp (string): `aqua >= v2.65.0`
- certificate_oidc_issuer (string): `aqua >= v2.65.0`
- certificate_github_workflow_repository (string): `aqua >= v2.65.0`
- certificate_github_workflow_ref (string): `aqua >= v2.65.0`

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

## Certificate fields

The fields added in v2.65.0 are passed to `cosign verify-blob` as the flag of the same name, after `opts`, and only when they are set. They are templates like `opts`, so a field can name every version's identity.

| field | flag |
| --- | --- |
| certificate_identity | `--certificate-identity` |
| certificate_identity_regexp | `--certificate-identity-regexp` |
| certificate_oidc_issuer | `--certificate-oidc-issuer` |
| certificate_github_workflow_repository | `--certificate-github-workflow-repository` |
| certificate_github_workflow_ref | `--certificate-github-workflow-ref` |

They say what `opts` can say too, but as fields, so what is verified can be read without parsing a command line. A flag shouldn't be given both ways.

```yaml
cosign:
  bundle:
    type: github_release
    asset: "{{.Asset}}.sigstore.json"
  certificate_identity: https://github.com/foundry-rs/foundry/.github/workflows/release.yml@refs/tags/{{.Version}}
  certificate_oidc_issuer: https://token.actions.githubusercontent.com
```
