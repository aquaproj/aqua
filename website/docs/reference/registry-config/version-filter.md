---
sidebar_position: 300
---

# version_filter

[v0.8.13](https://github.com/aquaproj/aqua/releases/tag/v0.8.13)

:::caution
In many cases you should use [version_prefix](version-prefix.md) rather than version_filter, because `version_prefix` makes Registry configuration simple.
:::

`aqua g` gets the latest version of the package.
If `version_filter` is set, the version which matches with `version_filter` is used.
`version_filter` is [expr](https://github.com/antonmedv/expr)'s expression.
The evaluation result must be a boolean.

This is used in `kubernetes-sigs/kustomize` to exclude [releases of kyaml](https://github.com/kubernetes-sigs/kustomize/releases?q=kyaml&expanded=true).

```yaml
- type: github_release
  repo_owner: kubernetes-sigs
  repo_name: kustomize
  asset: 'kustomize_{{trimPrefix "kustomize/" .Version}}_{{.OS}}_{{.Arch}}.tar.gz'
  version_filter: 'Version startsWith "kustomize/"'
```
