---
sidebar_position: 290
---

# vars

[v2.31.0](https://github.com/aquaproj/aqua/releases/tag/v2.31.0) [#3052](https://github.com/aquaproj/aqua/pull/3052)

e.g.

```yaml
packages:
  - type: github_release
    repo_owner: zoncoen
    repo_name: scenarigo
    asset: scenarigo_{{.Version}}_go{{.Vars.go_version}}_{{.OS}}_{{.Arch}}.{{.Format}}
    vars:
      - name: go_version
        required: true
    format: tar.gz
    replacements:
      amd64: x86_64
      darwin: Darwin
      linux: Linux
```

`vars` is a list of variables.
Fields of a variable:

- name: string (Required): A variable name
- required: boolean (Optional): If true, the variable is required. To use the package, users need to set the variable in aqua.yaml
- default: any (Optional): The default value of the variable

Variables are passed to template strings as `.Vars.<template name>`.

e.g.

```
asset: cpython-{{.Vars.python_version}}+{{.Version}}-{{.Arch}}-{{.OS}}-install_only.{{.Format}}
```

## `vars` in aqua.yaml

e.g.

```yaml
packages:
  - name: zoncoen/scenarigo@v0.17.3
    vars:
      go_version: 1.22.2
```

`vars` is a map of variables.
The key is a variable name and the value is a variable value.
