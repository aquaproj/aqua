---
sidebar_position: 300
---

# Fix the default `files[].name`

[#1409](https://github.com/aquaproj/aqua/issues/1409) [#1516](https://github.com/aquaproj/aqua/pull/1516)

If the package has a `name` field, the `name` is split with `/` and the last element is used as the default file name.

For example, please see the following example.

```yaml
name: cert-manager/cert-manager/cmctl
repo_owner: cert-manager
repo_name: cert-manager
```

Then in aqua v1 the default setting of `files` is the following.

```yaml
files:
- name: cert-manager
```

On the other hand, in aqua v2 the default setting of `files` is the following.

```yaml
files:
- name: cmctl
```

## Why this change is needed

We think aqua v2's default setting is better than aqua v1 in many cases.

## How to migrate

If you maintain registries, you might have to fix them.
