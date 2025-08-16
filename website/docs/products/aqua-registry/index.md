---
sidebar_position: 50
---

# aqua-registry

aqua's Standard Registry

https://github.com/aquaproj/aqua-registry

## Search packages from the Standard Registry by the command `aqua g`

Please add the Standard Registry to your aqua.yaml's registries, and run `aqua g`.

```yaml
registries:
  - type: standard
    ref: v4.233.0  # renovate: depName=aquaproj/aqua-registry
```

`aqua g` launches the interactive UI and you can search the package by fuzzy search.

```console
  civo/cli [civo]                          ┌──────────────────────────────────────────┐
  dapr/cli [dapr]                          │ climech/grit                             │
  goark/gimei-cli                          │                                          │
  orhun/git-cliff                          │ https://github.com/climech/grit          │
  snyk/cli [snyk]                          │ Multitree-based personal task manager    │
  spf13/cobra-cli                          │                                          │
  volta-cli/volta                          │                                          │
  barnybug/cli53                           │                                          │
  michidk/vscli                            │                                          │
  nuclio/nuclio                            │                                          │
  sigi-cli/sigi                            │                                          │
  tektoncd/cli                             │                                          │
  cswank/kcli                              │                                          │
  cli/cli [gh]                             │                                          │
> climech/grit                             │                                          │
  225/1569                                 │                                          │
> cli                                      └──────────────────────────────────────────┘
```

## Request for new packages

Please check [registry.yaml](https://github.com/aquaproj/aqua-registry/blob/main/registry.yaml) or search packages by `aqua g` command.
If the packages you want aren't found, please create issues or send pull requests!

By adding various packages to the standard registry, aqua becomes more useful and attractive.
We need your contribution!

## Contributing

Please see [Contributing](contributing.md).

## :bulb: Tips: Get all packages in your laptop

[Install Standard Registry's all packages very quickly](/docs/guides/install-all-packages)
