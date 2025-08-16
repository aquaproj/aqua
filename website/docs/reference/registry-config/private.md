---
sidebar_position: 2200
---

# private

aqua >= [v1.27.0](https://github.com/aquaproj/aqua/releases/tag/v1.27.0)

[#1466](https://github.com/aquaproj/aqua/issues/1466) [#1468](https://github.com/aquaproj/aqua/issues/1468) [#1486](https://github.com/aquaproj/aqua/pull/1486)

You can set `private: true` to private packages and registries.
By default, `private` is `false`.
If `private` is true, aqua skips sending HTTP requests to download assets, because the requests always fail.
Even if the value of `private` attribute is wrong, you can install the registry and package.

e.g. aqua.yaml

```yaml
registries:
- name: foo
  type: github_content
  repo_owner: aquaproj
  repo_name: private-repository
  ref: v3.90.0
  private: true # Private Registry
  path: registry.yaml
```

e.g. registry.yaml

```yaml
packages:
# init: a
- type: github_release
  repo_owner: aquaproj
  repo_name: private-repository
  asset: 'foo_{{trimV .Version}}_{{title .OS}}_{{.Arch}}.tar.gz'
  private: true # Private Package
```
