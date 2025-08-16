---
sidebar_position: 900
---

# `http` Package

The package is downloaded from the specified URL.

e.g. [hashicorp/terraform](https://github.com/aquaproj/aqua-registry/blob/main/pkgs/hashicorp/terraform/registry.yaml)

```yaml
packages:
  - type: http
    repo_owner: hashicorp
    repo_name: terraform
    url: https://releases.hashicorp.com/terraform/{{trimV .Version}}/terraform_{{trimV .Version}}_{{.OS}}_{{.Arch}}.zip
    description: Terraform enables you to safely and predictably create, change, and improve infrastructure. It is an open source tool that codifies APIs into declarative configuration files that can be shared amongst team members, treated as code, edited, reviewed, and versioned
```

## Required fields

* type
* url: The template string of URL where the package is downloaded
