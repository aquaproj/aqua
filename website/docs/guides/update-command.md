---
sidebar_position: 300
---

# Update registries and packages by update command

:::info
You can update them by Renovate too.
Please see [Update packages by Renovate](renovate.md) too.
:::

[#1657](https://github.com/aquaproj/aqua/issues/1657) [#2329](https://github.com/aquaproj/aqua/pull/2329) aqua >= [v2.14.0](https://github.com/aquaproj/aqua/releases/tag/v2.14.0)

You can update registries and packages by `aqua update (up)` command.

If no argument is passed, all registries and packages are updated to the latest.

```sh
# Update all packages and registries to the latest versions
aqua update
```

This command has an alias "up"

```sh
aqua up
```

This command

- gets the latest version from GitHub Releases, GitHub Tags, and crates.io and updates aqua.yaml
- doesn't install packages

## Updated configuration file paths

This command finds a configuration file `aqua.yaml` according to [the rule](/docs/reference/config/#configuration-file-path) and updates only one file.
Once this command finds one file, it stops searching other aqua.yaml.

So if you want to update other files, please change the current directory or specify the configuration file path with the option `-c`.

```sh
aqua -c foo/aqua.yaml update
```

## Update only registries

If you want to update only registries, please use the `--only-registry [-r]` option.

```sh
# Update only registries
aqua update -r
```

## Update only packages

If you want to update only packages, please use the `--only-package [-p]` option.

```sh
# Update only packages
aqua update -p
```

## Update only specific commands

You can specify packages with command names. aqua finds packages that have these commands and updates them.

```sh
aqua update <command name> [<command name> ...]
```

e.g.

```sh
# Update cli/cli
aqua update gh
```

## Specify versions

You can specify versions.

```sh
aqua update gh@v2.30.0
```

## Select packages with Fuzzy Finder

If you want to update only specific packages, please use the `-i` option.
You can select packages with the fuzzy finder.
If `-i` option is used, registries aren't updated.

```sh
# Select updated packages with fuzzy finder
aqua update -i
```

## Select the package version with Fuzzy Finder

If you want to select versions, please use the `-s` option.
You can select versions with the fuzzy finder. You can not only update but also downgrade packages.

```sh
# Select updated packages and versions with fuzzy finder
aqua update -i -s
```

## The field `version` is ignored

This command doesn't update packages if the `version` field is used.

```yaml
packages:
  - name: cli/cli@v2.0.0 # Update
  - name: gohugoio/hugo
    version: v0.118.0 # Doesn't update
```

So if you don't want to update specific packages, the `version` field is useful.

## commit hashes are ignored

This command doesn't update commit hashes.

```yaml
packages:
  - name: google/pprof@d04f2422c8a17569c14e84da0fae252d9529826b # Doesn't update
```

## Exclude some packages from the target of `aqua update`

aqua >= [v2.25.0](https://github.com/aquaproj/aqua/releases/tag/v2.25.0) [#2749](https://github.com/orgs/aquaproj/discussions/2749#discussioncomment-8808062) [#2752](https://github.com/aquaproj/aqua/pull/2752)

`aqua update` updates all packages by default, but you may not want to update some packages.
In this case, you can exclude some packages from the target of `aqua update`.

e.g. aqua.yaml

```yaml
packages:
- name: golang/vuln/govulncheck@v1.0.3
  update:
    # If enabled is false, aqua up command ignores the package.
    # If the package name is passed to aqua up command explicitly, enabled is ignored.
    # By default, enabled is true.
    enabled: false
```

If you specify the package explicitly, the setting `enabled` is ignored.

e.g.

```console
$ aqua up govuluncheck # the package is updated even if update.enabled is false
```

## Known Issues

There are some known issues related to the third party library [goccy/go-yaml](https://github.com/goccy/go-yaml).

### `null` is set to `packages` wrongly if registries are updated and `packages` is empty

This issue is because of the third party library [goccy/go-yaml](https://github.com/goccy/go-yaml).

Before

```yaml
registries:
- ref: v4.155.0
  type: standard
packages:
```

Run `aqua up`.

```console
$ aqua up
INFO[0000] updating a registry                           aqua_version= env=darwin/arm64 new_version=v4.155.1 old_version=v4.155.0 program=aqua registry_name=standard
```

After

```yaml
registries:
- ref: v4.155.1
  type: standard
packages: null
```
