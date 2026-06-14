---
sidebar_position: 550
---

# Security

We think security is very important and are working on improving the security of aqua.
aqua should allow you to install and execute tools securely.
In this page, we describe aqua's security perspective.

## List of Issues and Pull Requests

https://github.com/search?q=org%3Aaquaproj+label%3Asecurity

## Features

- Design
  - aqua doesn't execute external commands except for `go install` and `go build` to install packages
    - This prevents malicious commands from being executed
  - Centrally managed Registry is provided
    - Compared with third party registries, it has low risk to be tampered
- [Checksum Verification](checksum.md)
- [Policy as Code](policy-as-code/index.md)
  - [Only standard registry is allowed by default (Secure by default)](/docs/reference/upgrade-guide/v2/only-standard-registry-is-allowed-by-default)
- [ghtkn integration](ghtkn.md)
- [Cosign and SLSA Provenance](cosign-slsa.md)
- [Minisign](minisign.md)
- [GitHub Artifact Attestations](github-artifact-attestations.md)
- [GitHub Immutable Releases](github-immutable-release.md)
  - :warning: This feature was abandoned at aqua v2.60.1 [#4862](https://github.com/aquaproj/aqua/issues/4862)
- [Manage a GitHub access token using Keyring](keyring.md)
