---
sidebar_position: 310
---

# version_prefix

[#1545](https://github.com/aquaproj/aqua/issues/1545)

aqua >= [v1.31.0](https://github.com/aquaproj/aqua/releases/tag/v1.31.0)

You can filter versions with a specific prefix and trim the prefix from versions by `version_prefix`.

For example, [kubernetes-sigs/kustomize](https://github.com/kubernetes-sigs/kustomize/releases?q=kustomize%2F&expanded=true) has a prefix `kustomize/`.

```yaml
- type: github_release
    repo_owner: kubernetes-sigs
    repo_name: kustomize
    version_prefix: kustomize/
    asset: kustomize_{{.SemVer}}_{{.OS}}_{{.Arch}}.tar.gz
    version_constraint: semver(">= 4.5.4")
    version_overrides:
      - version_constraint: semver(">= 4.4.1")
        supported_envs:
          - linux
          - darwin
          - amd64
      - version_constraint: semver(">= 4.2.0")
        supported_envs:
          - linux
          - darwin
      - version_constraint: semver("< 4.2.0")
        rosetta2: true
        supported_envs:
          - linux
          - darwin
          - amd64
```
