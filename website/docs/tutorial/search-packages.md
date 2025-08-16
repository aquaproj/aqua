---
sidebar_position: 200
---

# Search packages

You can search packages by [aqua g](/docs/reference/usage#aqua-generate) command.

```bash
aqua g
```

Then an interactive fuzzy zinder is launched (Powered by [ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder)).

```
  okta/okta-aws-cli                     ┌──────────────────────────────────────┐
  openfaas/faas-cli                     │ climech/grit                         │
  yitsushi/totp-cli                     │                                      │
  databricks/click                      │ https://github.com/climech/grit      │
  ipinfo/cli/prips                      │ Multitree-based personal task manag..│
  rgreinho/tfe-cli                      │                                      │
  civo/cli [civo]                       │                                      │
  dapr/cli [dapr]                       │                                      │
  goark/gimei-cli                       │                                      │
  orhun/git-cliff                       │                                      │
  snyk/cli [snyk]                       │                                      │
  spf13/cobra-cli                       │                                      │
  volta-cli/volta                       │                                      │
  barnybug/cli53                        │                                      │
  michidk/vscli                         │                                      │
  nuclio/nuclio                         │                                      │
  sigi-cli/sigi                         │                                      │
  cswank/kcli                           │                                      │
  cli/cli [gh]                          │                                      │
> climech/grit                          │                                      │
  191/1303                              │                                      │
> cli                                   └──────────────────────────────────────┘
```

Please select `tfmigrator/cli`, then the package configuration is outputted.

```console
$ aqua g
- name: tfmigrator/cli@v0.2.2
```

You can select multiple packages by tab key.

If `-i` option is set, then the package is added to `aqua.yaml`.

```bash
aqua g -i
```

```yaml
packages:
- name: cli/cli@v2.38.0
- name: junegunn/fzf@0.43.0
- name: tfmigrator/cli@v0.2.2 # Added
```
