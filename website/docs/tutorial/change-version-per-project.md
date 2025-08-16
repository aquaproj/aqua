---
sidebar_position: 500
---

# Use different version per project

aqua supports changing the tool versions per project.

```bash
mkdir foo bar
echo -n 'registries:
- type: standard
  ref: v4.79.0 # renovate: depName=aquaproj/aqua-registry

packages:
- name: cli/cli@v2.1.0
' > foo/aqua.yaml
echo -n 'registries:
- type: standard
  ref: v3.79.0 # renovate: depName=aquaproj/aqua-registry

packages:
- name: cli/cli@v2.0.0
' > bar/aqua.yaml
```

```bash
cd foo
gh version # In foo, the version is v2.1.0.
cd ../bar
gh version # In bar, the version is v2.0.0.
```

The version of `gh` is changed seamlessly.

You can install multiple versions in the same laptop and switch the version by project.
