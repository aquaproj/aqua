---
sidebar_position: 10
---

# Update packages by Renovate

```json
{
  "extends": [
    "github>aquaproj/aqua-renovate-config#2.2.1"
  ]
}
```

![image](https://user-images.githubusercontent.com/13323303/176582627-44f27c48-213b-44da-b18f-d4d482ef2f56.png)

:::info
As of aqua v2.14.0, you can update them by `aqua update` command too.
Please see [Update registries and packages by update command](update-command.md) too.
:::

aqua manages package and registry versions,
so it is important to update them continuously.
We recommend managing `aqua.yaml` with Git and update versions by [Renovate](https://docs.renovatebot.com/) or something.

Using Renovate's [Regex Manager](https://docs.renovatebot.com/modules/manager/regex/), you can update versions.

We provide the Renovate Preset Configuration [aqua-renovate-config](https://github.com/aquaproj/aqua-renovate-config). [For details, please see the document](/docs/products/aqua-renovate-config).

Example pull requests by Renovate.

* [chore(deps): update dependency golangci/golangci-lint to v1.42.0](https://github.com/aquaproj/aqua/pull/193)
* [chore(deps): update dependency aquaproj/aqua-registry to v0.2.2](https://github.com/aquaproj/aqua/pull/194)

## :bulb: Use Renovate with Dependabot

If you already use Dependabot and hesitate to use Renovate, you should enable only Renovate's `regex` Manager.

```json
{
  "enabledManagers": ["regex"]
}
```

Then Renovate doesn't conflict with Dependabot.

## :bulb: Schedule Standard Registry's update

The release frequency of [Standard Registry](https://github.com/aquaproj/aqua-registry) is high.
If you feel the update of Standard Registry is a bit noisy, you can schedule the update of Standard Registry.

- [schedule | Renovate](https://docs.renovatebot.com/configuration-options/#schedule)
- [Schedule Presets | Renovate](https://docs.renovatebot.com/presets-schedule/)

e.g.

```json
{
  "packageRules": [
    {
      "matchPackageNames": ["aquaproj/aqua-registry"],
      "extends": ["schedule:earlyMondays"]
    }
  ]
}
```

## :bulb: Renovate's minimumReleaseAge

Some packages have a time lag between when a GitHub Release is created and when assets are uploaded.
In that case, Renovate may create pull requests before assets are uploaded.
To prevent the issue as much as possible, Renovate's [minimumReleaseAge](https://docs.renovatebot.com/configuration-options/#minimumreleaseage) may be useful.

## :bulb: Update aqua-checksums.json

aqua-renovate-config can't update aqua-checksums.json.
There are several ways to update aqua-checksums.json:

1. [Set up CI. Please see the guide.](./checksum.md)
1. Run `aqua upc` command via [postUpgradeTasks](https://docs.renovatebot.com/configuration-options/#postupgradetasks)

For details of postUpgradeTasks, please see the document of Renovate.

## :bulb: Prevent some packages from being updated by Renovate

There are two ways to prevent some packages from being updated by Renovate.

1. [Renovate's enabled option](https://docs.renovatebot.com/configuration-options/#enabled)
2. Use the long syntax instead of the short syntax

### 1. Renovate's enabled option

e.g. renovate.json

```json
{
  "packageRules": [
    {
      "matchPackageNames": ["kubernetes/kubectl"],
      "enabled": false
    }
  ]
}
```

### 2. Use the long syntax instead of the short syntax

e.g. aqua.yaml

:thumbsup: Renovate wouldn't update `kubernetes/kubectl`.

```yaml
packages:
- name: kubernetes/kubectl
  version: v1.25.0
```

:thumbsdown: Renovate would update `kubernetes/kubectl` and `suzuki-shunsuke/tfcmt`.

```yaml
packages:
- name: kubernetes/kubectl@v1.25.0
- name: suzuki-shunsuke/tfcmt
  version: v2.0.0 # renovate: depName=suzuki-shunsuke/tfcmt
```
