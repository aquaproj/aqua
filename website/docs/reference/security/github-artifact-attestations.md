---
sidebar_position: 1260
---

# GitHub Artifact Attestations

- `aqua >= v2.35.0` [#3119](https://github.com/aquaproj/aqua/pull/3119)

You can verify packages' [GitHub Artifact Attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds) if they are provided.

## How to verify packages

You don't have to do any special things.
If packages provide GitHub Artifact Attestations and registries are configured, packages are verified when you install them.
aqua uses [GitHub CLI](https://cli.github.com/) internally, but aqua installs it in `$(aqua root-dir)` automatically, so you don't have to install it.

e.g.

```yaml
# aqua.yaml
checksum:
  enabled: true
registries:
  - type: standard
    ref: f9cce37273a70e2f5664fb4c3708169ffe7e320c # TODO update
packages:
  - name: suzuki-shunsuke/mkghtag@v0.1.5-3
```

mkghtag provides GitHub Artifact Attestations.

https://github.com/suzuki-shunsuke/mkghtag/attestations

When you install mkghtag, GitHub Artifact Attestations are verified.

```console
$ aqua i
...
INFO[0000] verify GitHub Artifact Attestations           aqua_version= env=darwin/arm64 exe_name=mkghtag package_name=suzuki-shunsuke/mkghtag package_version=v0.1.5-3 program=aqua registry=standard
Loaded digest sha256:5e79e447d4a664da3fbed12c6486bfb18eeb846d9aceb7b6eaa42277f04dcf6b for file:///var/folders/fc/1bgyy3_d3x90m_t04qbw5f8m0000gn/T/767459864
Loaded 1 attestation from GitHub API
✓ Verification succeeded!

sha256:5e79e447d4a664da3fbed12c6486bfb18eeb846d9aceb7b6eaa42277f04dcf6b was attested by:
REPO                                 PREDICATE_TYPE                  WORKFLOW                                                                   
suzuki-shunsuke/go-release-workflow  https://slsa.dev/provenance/v1  .github/workflows/release.yaml@refs/heads/feat-github-artifact-attestations
...
```

### Disable the verification of GitHub Artifact Attestations

We recommend enabling the verification for security, but you can disable the verification by the environment variable.

```sh
export AQUA_DISABLE_GITHUB_ARTIFACT_ATTESTATION=true
```

## Registry Settings

e.g.

```yaml
packages:
  - type: github_release
    repo_owner: suzuki-shunsuke
    repo_name: mkghtag
    asset: mkghtag_{{.OS}}_{{.Arch}}.{{.Format}}
    format: tar.gz
    github_artifact_attestations: # asset's GitHub Artifact Attestations
      signer-workflow: suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
    checksum:
      type: github_release
      asset: mkghtag_{{trimV .Version}}_checksums.txt
      algorithm: sha256
      github_artifact_attestations: # checksum's GitHub Artifact Attestations
        signer-workflow: suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```
