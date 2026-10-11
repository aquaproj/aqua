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
- repository (string): `aqua >= v2.65.0`
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

Each of the fields added in v2.65.0 is passed to `gh attestation verify` as the flag below, and only when it is set:

| field | flag | what it checks |
| --- | --- | --- |
| repository | `--repo` | the repository the artifact is attested in, as `owner/name`. It defaults to the package's repository, so it is needed only when the attestation is in another repository |
| signer_repo | `--signer-repo` | the repository of the workflow that signed, without pinning the workflow's path |
| cert_oidc_issuer | `--cert-oidc-issuer` | the OIDC issuer of the signing certificate. gh defaults to GitHub Actions' issuer `https://token.actions.githubusercontent.com`, so it only needs to be set for another issuer |
| deny_self_hosted_runners | `--deny-self-hosted-runners` | that the artifact wasn't built on a self-hosted runner |
| source_ref | `--source-ref` | the ref the artifact was built from, such as `refs/tags/v1.2.3` |
| source_digest | `--source-digest` | the commit the artifact was built from |

`signer_repo` is a looser alternative to `signer_workflow`, not an addition to it: gh accepts only one of them, and aqua fails without running gh when both are set. `signer_workflow` already names the repository, including a reusable workflow's in another repository.

`source_ref` and `source_digest` differ for every version, so they are for a configuration that describes one version, such as an entry of aqua-registry-g2.

```yaml
github_artifact_attestations:
  signer_workflow: suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
  deny_self_hosted_runners: true
```

`github_artifact_attestations` in `version_overrides` replaces the whole setting rather than merging into it, so a field set at the top level, such as `deny_self_hosted_runners`, has to be repeated in each override that sets `github_artifact_attestations`.
