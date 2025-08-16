---
sidebar_position: 1300
---

# `go_install` Package

[#823](https://github.com/aquaproj/aqua/issues/823) [#826](https://github.com/aquaproj/aqua/pull/826), aqua >= [v1.10.0](https://github.com/aquaproj/aqua/releases/tag/v1.10.0) is required.

* `path`: Go package path. If `path` is not set but `repo_owner` and `repo_name` are set, the package path is `github.com/<repo_owner>/<repo_name>`
* `name`: The package name. If `name` is not set but `repo_owner` and `repo_name` are set, the package name is `<repo_owner>/<repo_name>`. If `name`, `repo_owner`, and `repo_name` aren't set, `path` is used as the package name

The package is installed by `go install` command.
So the command `go` is required.
When aqua runs `go install`, aqua sets the environment variable `GOBIN`.

aqua is a CLI Version Manager, you have to specify the version. You can't specify `latest`.

e.g. [golang.org/x/perf/cmd/benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)

registry.yaml

```yaml
packages:
  - type: go_install
    path: golang.org/x/perf/cmd/benchstat
    description: Benchstat computes and compares statistics about benchmarks
```

aqua.yaml

```yaml
registries:
  - name: local
    type: local
    path: registry.yaml
packages:
  - name: golang.org/x/perf/cmd/benchstat
    registry: local
    version: 84e58bfe0a7e5416369e236afa007d5d9c58a0fa
```

```console
$ aqua i
INFO[0000] create a symbolic link                        aqua_version= env=darwin/arm64 link_file=/Users/shunsukesuzuki/.local/share/aquaproj-aqua/bin/benchstat new=aqua-proxy package_name=golang.org/x/perf/cmd/benchstat package_version=84e58bfe0a7e5416369e236afa007d5d9c58a0fa program=aqua registry=local
INFO[0000] Installing a Go tool                          aqua_version= env=darwin/arm64 go_package_path=golang.org/x/perf/cmd/benchstat@84e58bfe0a7e5416369e236afa007d5d9c58a0fa gobin=/Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/go_install/golang.org/x/perf/cmd/benchstat/84e58bfe0a7e5416369e236afa007d5d9c58a0fa/bin package_name=golang.org/x/perf/cmd/benchstat package_version=84e58bfe0a7e5416369e236afa007d5d9c58a0fa program=aqua registry=local
go: downloading golang.org/x/perf v0.0.0-20220411212318-84e58bfe0a7e

$ aqua which benchstat
/Users/shunsukesuzuki/.local/share/aquaproj-aqua/pkgs/go_install/golang.org/x/perf/cmd/benchstat/84e58bfe0a7e5416369e236afa007d5d9c58a0fa/bin/github-compare
```

[#1084](https://github.com/aquaproj/aqua/issues/1084) [#1487](https://github.com/aquaproj/aqua/pull/1487) From aqua [v1.27.0](https://github.com/aquaproj/aqua/releases/tag/v1.27.0), `path` is treated as a template string.

e.g.

```yaml
packages:
  - type: go_install
    repo_owner: volatiletech
    repo_name: sqlboiler
    description: Generate a Go ORM tailored to your database schema
    path: github.com/volatiletech/sqlboiler/v{{(semver .Version).Major}}
```
