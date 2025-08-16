---
sidebar_position: 800
---

# `github_release` Package

The package is downloaded from GitHub Releases.

e.g. [suzuki-shunsuke/tfcmt](https://github.com/aquaproj/aqua-registry/blob/main/pkgs/suzuki-shunsuke/tfcmt/registry.yaml)

```yaml
packages:
  - type: github_release
    repo_owner: suzuki-shunsuke
    repo_name: tfcmt
    asset: tfcmt_{{.OS}}_{{.Arch}}.tar.gz
    description: Fork of mercari/tfnotify. tfcmt enhances tfnotify in many ways, including Terraform >= v0.15 support and advanced formatting options
```

## Required fields

* type
* repo_owner
* repo_name
* asset: The template string of GitHub Release's asset name
