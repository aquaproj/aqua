---
sidebar_position: 550
---

# Filter packages with tags

`aqua >= v1.23.0`

[#441](https://github.com/aquaproj/aqua/issues/441) [#1336](https://github.com/aquaproj/aqua/pull/1336)

`aqua cp`, `aqua install`, and `aqua update` commands support filtering packages with tags.

## Specification

The optional field `tags` can be specified in `aqua.yaml`'s package.
This field is a string list of tags.

e.g.

```yaml
---
registries:
- type: standard
  ref: v4.155.1  # renovate: depName=aquaproj/aqua-registry
packages:
- name: suzuki-shunsuke/tfcmt@v3.2.0
  tags:
    - test
    - foo
- name: suzuki-shunsuke/github-comment@v4.0.0
- name: cli/cli@v2.0.0
  tags:
    - bar
    - foo
```

`cp`, `install`, and `update` commands have the following command line options.

- `--tags (-t)` (string): When this option is set, only packages that have specified tags are installed. You can specify multiple tags joining with `,` (e.g. `-t ci,test`)
- `--exclude-tags` (string): When this option is set, packages that have specified tags aren't installed. You can specify multiple tags joining with `,` (e.g. `-exclude-tags ci,test`)

In case of the above `aqua.yaml`, you can filter packages as the following.

```console
$ aqua i # Install suzuki-shunsuke/tfcmt@v3.2.0 and suzuki-shunsuke/github-comment@v4.0.0 and cli/cli@v2.0.0
$ aqua i -t test # Install suzuki-shunsuke/tfcmt@v3.2.0
$ aqua i -t foo,bar # Install suzuki-shunsuke/tfcmt@v3.2.0 and cli/cli@v2.0.0
$ aqua i --exclude-tags test # Install suzuki-shunsuke/github-comment@v4.0.0 and cli/cli@v2.0.0
$ aqua i --exclude-tags test -t foo # Install cli/cli@v2.0.0
```

:::caution
Note that `aqua install` creates symbolic links of all packages regardless tags, so that you can execute all tools by Lazy Install and assure that tools are managed by aqua.
:::
