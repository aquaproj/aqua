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
* build_tags: Go build tags passed to `go build -tags`

```
${AQUA_ROOT_DIR}/pkgs/go_build/github.com/google/wire/v0.5.0/
  bin/wire
  src/ # GitHub Repository Archive
    wire-0.5.0/ # `go build` is run on this directory
      cmd/wire # build target
```

## `build_tags`

Some Go programs need build tags to compile without external C libraries.
`build_tags` passes the tags to `go build -tags`.

e.g. https://github.com/podman-container-tools/skopeo

```yaml
packages:
  - type: go_build
    repo_owner: podman-container-tools
    repo_name: skopeo
    description: Work with remote image registries
    files:
      - name: skopeo
        src: ./cmd/skopeo
        dir: skopeo-{{trimV .Version}}
        build_tags:
          - containers_image_openpgp
          - exclude_graphdriver_btrfs
```

aqua joins the tags with a comma and runs:

```sh
go build -tags containers_image_openpgp,exclude_graphdriver_btrfs -o <exe_path> ./cmd/skopeo
```

Without these tags, skopeo imports `github.com/proglottis/gpgme`, which
requires the system library `libgpgme` and `pkg-config`. The tags remove the
dependency from the import graph, so the build needs no external library.

If `build_tags` is empty, aqua does not pass `-tags`.
