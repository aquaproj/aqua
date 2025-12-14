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

## GitHub Enterprise Server (GHES) Registry

`aqua >= v2.56.0`

You can use registries hosted on GitHub Enterprise Server.

See [GitHub Enterprise Server Support](/docs/guides/github-enterprise-server) for details.

## Create your own Registry

Please see the following document.

- [Develop a Registry](/docs/develop-registry/)
