---
sidebar_position: 250
---

# Registry

`Registry` is a core concept of `aqua`.
`Registry` defines the package list and how to install them.

In `aqua.yaml`, you specify Registries in `registries`.

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry

packages:
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
```

You don't have to define how to install tools yourself because it is defined in the Registry.
Registry is reusable across multiple configuration.

## Standard Registry

In the above configuration, the Standard Registry v2.22.0 is used.
The Standard Registry is a registry that we maintain.

https://github.com/aquaproj/aqua-registry

## Registry Types

aqua supports multiple registry types:

- **[Standard Registry](https://github.com/aquaproj/aqua-registry)**: The official aqua registry maintained by the community
- **[Local Registry](/docs/reference/config/#local-registry)**: Use a local file as a registry
- **[GitHub Content Registry](/docs/reference/config/#github_content-registry)**: Fetch registry from GitHub repository
- **[HTTP Registry](/docs/reference/config/#http-registry)** (aqua >= v2.56.0): Fetch registry from HTTP(S) endpoints for private internal registries

## Create your own Registry

Please see the following document.

- [Develop a Registry](/docs/develop-registry/)
