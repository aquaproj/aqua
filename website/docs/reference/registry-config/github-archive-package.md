---
sidebar_position: 1100
---

# `github_archive` Package

The package is downloaded from GitHub Archive.

e.g. [tfutils/tfenv](https://github.com/aquaproj/aqua-registry/blob/main/pkgs/tfutils/tfenv/registry.yaml)

```yaml
packages:
  - type: github_archive
    repo_owner: tfutils
    repo_name: tfenv
    description: Terraform version manager
    files:
      - name: tfenv
        src: tfenv-{{trimV .Version}}/bin/tfenv
      - name: terraform
        src: tfenv-{{trimV .Version}}/bin/terraform
```

## Required fields

* type
* repo_owner
* repo_name
