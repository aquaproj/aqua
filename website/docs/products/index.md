---
sidebar_position: 100
---

# Products

[Repositories](https://github.com/orgs/aquaproj/repositories)

- [aqua](https://github.com/aquaproj/aqua): Main product. CLI Version Manager
- [aqua-registry](aqua-registry/index.md): aqua's Standard Registry
- [aqua-installer](aqua-installer/index.md): A shell script and GitHub Actions to install aqua
- [circleci-orb-aqua](circleci-orb-aqua.md): CircleCI Orb to install aqua
- [aqua-renovate-config](aqua-renovate-config.md): Renovate Config Preset to update aqua, aqua-installer, packages, and registries
- [aquaproj.github.io](https://github.com/aquaproj/aquaproj.github.io): aqua's official website
- [devcontainer-features](devcontainer-features.md): [Dev Container Features](https://containers.dev/implementors/features/) for aqua

## Checksum Verification

[Checksum Verification](/docs/reference/security/checksum)

### GitHub Actions

- [update-checksum-action](update-checksum-action.md): GitHub Actions to update aqua-checksums.json. If aqua-checksums.json isn't latest, update aqua-checksums.json and push a commit
- [update-checksum-workflow](update-checksum-workflow.md): GitHub Actions Reusable Workflow to update aqua-checksums.json

### Example

- [example-update-checksum](https://github.com/aquaproj/example-update-checksum)
- [example-update-checksum-public](https://github.com/aquaproj/example-update-checksum-public)

## Develop Registry

[Develop Registry](/docs/develop-registry/)

- [registry-tool](https://github.com/aquaproj/registry-tool): 
- [generate-registry-deep](https://github.com/aquaproj/generate-registry-deep)

### Develop custom Registry

- [registry-action](https://github.com/aquaproj/registry-tool): 
- [template-private-aqua-registry](https://github.com/aquaproj/template-private-aqua-registry)

## Internal tools

- [aqua-proxy](aqua-proxy.md)

## Archived

- [slsa-verifier](https://github.com/aquaproj/slsa-verifier): Fork of [slsa-framework/slsa-verifier](https://github.com/slsa-framework/slsa-verifier)
