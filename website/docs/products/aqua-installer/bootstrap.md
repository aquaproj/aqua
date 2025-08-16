---
sidebar_position: 20
---

# Bootstrap

[#716](https://github.com/aquaproj/aqua-installer/issues/716)

aqua-installer installs a specific version of aqua before installing a given version of aqua.
We call this `bootstrap`.
This page describes why `bootstrap` is necessary and aqua-installer doesn't install a give version directly.

Bootstrap is required to achieve secure install easily with minimum dependency.

- Secure Install: verify checksums and signature
- Minimum dependency: aqua-installer requires only curl or wget
- Easy implementation: aqua-installer lets complicated process to `aqua update-aqua`

[We think Supply Chain Security is so important](/docs/reference/security/).

To verify checksums, you need to get expected checksums somehow.
aqua-installer supports any versions, so it needs to be able to get checksums of any versions.
Downloaded checksum files must be verified somehow.
aqua uses tools such as [Cosign](https://github.com/sigstore/cosign) for verification, so they must be installed securely before installing aqua.
To install Cosign securely, Cosign must be verified. This is a bootstrap issue.

To solve this issue, we hardcode a specific aqua version and checksums in aqua-installer.
We can trust hardcoded checksums.
aqua-installer downloads the specific version and verifies checksums without Cosign.
Then aqua-installer executes `aqua update-aqua` command to install Cosign and aqua securely.
We hardcode Cosign version and checksums in aqua to install Cosign securely.
`aqua update-aqua` command installs Cosign securely, and installs aqua securely.

That's why the bootstrap is required.

Without bootstrap, you need to install one of the following tools before executing aqua-installer, which making aqua-installer hard to use.

- Cosign: verification of checksum files
- [slsa-verifier](https://github.com/slsa-framework/slsa-verifier): verification of SLSA Provenance
- [GitHub CLI and GitHub Access Token: verification of GitHub Artifact Attestations](https://github.blog/changelog/2024-06-25-artifact-attestations-is-generally-available/)

And we need to take care of compatibility. For example, users may use old GitHub CLI not supporting [gh attestation](https://cli.github.com/manual/gh_attestation) command.

We hardcode versions and checksums of tools which aqua uses internally not to depend on user environment.

aqua-installer lets complicated process to `aqua update-aqua`, making aqua-installer simple.

We accept some overhead due to bootstrap.
