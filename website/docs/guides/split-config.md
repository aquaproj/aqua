---
sidebar_position: 200
---

# Split the list of packages

You can split the list of packages.

e.g.

Directory structure

```
aqua.yaml
aqua/
  conftest.yaml
```

aqua.yaml

```yaml
registries:
- type: standard
  ref: v4.155.1  # renovate: depName=aquaproj/aqua-registry

packages:
- import: aqua/*.yaml
```

aqua/conftest.yaml

```yaml
packages:
- name: open-policy-agent/conftest@v0.28.2
```

This is useful for CI.
You can execute test and lint only when the specific package is updated.

e.g. GitHub Actions' [`on.<push|pull_request>.paths`](https://docs.github.com/en/actions/learn-github-actions/workflow-syntax-for-github-actions#onpushpull_requestpaths)

```yaml
name: conftest
on:
  pull_request:
    paths:
    - policy/**.rego
    - aqua/conftest.yaml
```

## :bulb: Renovate Config Preset

To update split files by Renovate, please use the preset [aquaproj/aqua-renovate-config:file](https://github.com/aquaproj/aqua-renovate-config#file-preset)

e.g.

```json
{
  "extends": [
    "github>aquaproj/aqua-renovate-config:file#2.2.1(aqua/conftest\\.yaml)"
  ]
}
```

You can also use the regular expression.

```json
{
  "extends": [
    "github>aquaproj/aqua-renovate-config:file#2.2.1(aqua/.*\\.ya?ml)"
  ]
}
```

## `import_dir`

[#3528](https://github.com/aquaproj/aqua/pull/3528) aqua >= v2.44.0

You can also use the `import_dir` field.

e.g. aqua.yaml

```yaml
registries:
- type: standard
  ref: v4.311.0
import_dir: imports
```

```sh
aqua init -u # Create aqua.yaml with `import_dir: imports`
aqua init -i pkgs # Create aqua.yaml with `import_dir: pkgs`
```

You can use `import_dir` and `packages` at the same time.
In addition to `packages`, aqua searches packages from the directory specified with `import_dir`.

`import_dir: imports` is equivalent to the following settings.

```yaml
packages:
- import: imports/*.yml
- import: imports/*.yaml
```

And if `import_dir` is set, `aqua g -i` command creates a directory `<import_dir>` and adds packages to the file `<import_dir>/<command name>.yaml`.
For instance, if `import_dir` is `imports`, `aqua g -i cli/cli` creates a directory `imports` and adds cli/cli to `imports/gh.yaml`.

If the package has multiple commands, `<command name>` is the first command name in the `files` setting.
For instance, in case of `FiloSottile/age`, `<command name>` is `age`.

https://github.com/aquaproj/aqua-registry/blob/d39d4b5d0fb0635f6be7a70f3cb8b994f075a639/pkgs/FiloSottile/age/registry.yaml#L13-L17

```yaml
    files:
      - name: age
        src: age/age
      - name: age-keygen
        src: age/age-keygen
```
