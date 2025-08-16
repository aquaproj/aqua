---
sidebar_position: 30
---

# Keep configuration files in one directory

- aqua >= v1.33.0 [#1595](https://github.com/aquaproj/aqua/issues/1595)
- [update-checksum-action](https://github.com/aquaproj/update-checksum-action) >= v0.2.0
- [update-checksum-workflow](https://github.com/aquaproj/update-checksum-workflow) >= v0.1.1

aqua has several configuration files such as `aqua.yaml`, [aqua-checksums.json](checksum.md), [policy file](/docs/reference/security/policy-as-code), and [imported files](split-config.md).

e.g.

```
aqua.yaml
aqua-checksums.json
aqua-policy.yaml # You can rename freely
aqua/ # You can rename freely
  terraform.yaml
  ...
```

From aqua v1.33.0, aqua supports keeping configuration files in one directory.

e.g.

```
aqua/ # or .aqua
  aqua.yaml
  aqua-checksums.json
  policy.yaml # You can rename freely
  imports/ # You can rename freely
    terraform.yaml
    ...
```

This is useful to keep the working directory clean.

## How to migrate

- Update aqua to v1.33.0 or later
- Update [update-checksum-action](https://github.com/aquaproj/update-checksum-action) to v0.2.0 or later
- Update [update-checksum-workflow](https://github.com/aquaproj/update-checksum-workflow) to v0.1.1 or later
- Move files
  - aqua.yaml => aqua/aqua.yaml
  - aqua-checksums.json => aqua/aqua-checksums.json
  - aqua-policy.yaml => aqua/policy.yaml # rename freely
  - [imported files](split-config.md) => aqua/imports/*.yaml # change the filename freely
- Fix [AQUA_POLICY_CONFIG](/docs/reference/security/policy-as-code)
- Fix GitHub Actions Path filters
- Fix [aqua-renovate-config](https://github.com/aquaproj/aqua-renovate-config)'s file preset argument
