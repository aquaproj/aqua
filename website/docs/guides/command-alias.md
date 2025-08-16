---
sidebar_position: 310
---

# Command Aliases

[v2.37.0](https://github.com/aquaproj/aqua/releases/tag/v2.37.0) [3224](https://github.com/aquaproj/aqua/pull/3224)

You can define command aliases in aqua.yaml.

This is useful to use multiple versions of the same package.

```yaml
registries:
- type: standard
  ref: v4.246.0 # renovate: depName=aquaproj/aqua-registry
packages:
- name: hashicorp/terraform@v1.9.8
- name: hashicorp/terraform
  version: v0.13.7
  command_aliases:
    - command: terraform
      alias: terraform-013
      # no_link: true
```

Then you can run `terraform` (v1.9.8) and `terraform-013` (v0.13.7).

```console
$ terraform version
Terraform v1.9.8
on darwin_arm64

$ terraform-013 version
Terraform v0.13.7

Your version of Terraform is out of date! The latest version
is 1.9.8. You can update by downloading from https://www.terraform.io/downloads.html
```

You can skip creating symbolic links for aliases by `no_link: true`

```yaml
  command_aliases:
    - command: terraform
      alias: terraform-013
      no_link: true
```

You can still run aliases via `aqua exec`.

```sh
aqua exec -- terraform-013 version
```
