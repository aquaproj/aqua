---
sidebar_position: 1250
---

# `go_build` Package

[#2131](https://github.com/aquaproj/aqua/issues/2131) [#2065](https://github.com/aquaproj/aqua/pull/2065)

* `repo_owner`: The repository owner name
* `repo_name`: The repository name

The repository archive is downloaded from GitHub and the package is built by command `go build` when it is installed.
So the command `go` is required.
aqua is a CLI Version Manager, you have to specify the version. Unlike `go install` command, you can't specify the head of the default branch.

e.g. https://github.com/google/wire

registry.yaml

```yaml
packages:
  - type: go_build
    repo_owner: google
    repo_name: wire
    description: Compile-time Dependency Injection for Go
    files:
      - name: wire
        src: ./cmd/wire
        dir: wire-{{trimV .Version}}
```

aqua.yaml

```yaml
registries:
  - name: local
    type: local
    path: registry.yaml
packages:
  - name: google/wire@v0.5.0
    registry: local
```

## File parameter

```yaml
    files:
      - name: wire
        src: ./cmd/wire
        dir: wire-{{trimV .Version}}
```

* name: command name
* dir: Directory path where `go build` is run
* src: go build's target path

```
${AQUA_ROOT_DIR}/pkgs/go_build/github.com/google/wire/v0.5.0/
  bin/wire
  src/ # GitHub Repository Archive
    wire-0.5.0/ # `go build` is run on this directory
      cmd/wire # build target
```
