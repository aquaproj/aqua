---
sidebar_position: 2120
---

# github_artifact_attestations

- `aqua >= v2.35.0`

Please see [Reference](/docs/reference/security/github-artifact-attestations) too.

## Fields

- enabled (bool)
- signer_workflow (string)
- predicate_type (string)
- signer_repo (string): `aqua >= v2.65.0`
- cert_oidc_issuer (string): `aqua >= v2.65.0`
- deny_self_hosted_runners (bool): `aqua >= v2.65.0`
- source_ref (string): `aqua >= v2.65.0`
- source_digest (string): `aqua >= v2.65.0`

e.g.

```yaml
github_artifact_attestations:
  signer_workflow: suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```

Each of the fields added in v2.65.0 is passed to `gh attestation verify` as the flag of the same name, and only when it is set:

| field | flag | what it checks |
| --- | --- | --- |
| signer_repo | `--signer-repo` | the repository of the workflow that signed, for a reusable workflow in another repository than the package's |
| cert_oidc_issuer | `--cert-oidc-issuer` | the OIDC issuer of the signing certificate, such as `https://token.actions.githubusercontent.com` |
| deny_self_hosted_runners | `--deny-self-hosted-runners` | that the artifact wasn't built on a self-hosted runner |
| source_ref | `--source-ref` | the ref the artifact was built from, such as `refs/tags/v1.2.3` |
| source_digest | `--source-digest` | the commit the artifact was built from |

`source_ref` and `source_digest` differ for every version, so they are for a configuration that describes one version, such as an entry of aqua-registry-g2.

```yaml
github_artifact_attestations:
  signer_workflow: suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
  signer_repo: suzuki-shunsuke/go-release-workflow
  cert_oidc_issuer: https://token.actions.githubusercontent.com
  deny_self_hosted_runners: true
```
