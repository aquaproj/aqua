# How does Lazy Install works?

In this document we describe how the Lazy Install works.
The Lazy Install is the aqua's characteristic feature, and maybe you feel it like magic.

By `aqua i`, aqua installs [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy) regardless the aqua's configuration.

```
~/.aqua/
  bin/
    aqua-proxy -> ../pkgs/github_release/github.com/suzuki-shunsuke/aqua-proxy/v0.1.2/aqua-proxy_darwin_amd64.tar.gz/aqua-proxy
  pkgs/
    github_release/github.com/suzuki-shunsuke/aqua-proxy/v0.1.2/aqua-proxy_darwin_amd64.tar.gz/aqua-proxy
```

And by `aqua i`, aqua reads the configuration file and creates symbolic links to aqua-proxy in `~/.aqua/bin`.
The symbolic link name is the package's file name.

For example, by the following configuration symbolic links `go` and `gofmt` are created.

```yaml
inline_registry:
  packages:
  - name: go
    type: http
    url: https://golang.org/dl/go{{.Version}}.{{.OS}}-{{.Arch}}.tar.gz
    files:
    - name: go # the symbolic `go` is created
      src: go/bin/go
    - name: gofmt # the symbolic `gofmt` is created
      src: go/bin/gofmt

packages:
- name: go
  registry: inline
  version: "1.17"
```

```
~/.aqua/
  bin/
    go -> aqua-proxy
    gofmt -> aqua-proxy
```

Add `~/.aqua/bin` to the environment variable `PATH`.
When `go version` is executed, `~/.aqua/bin/go` is a symbolic link to aqua-proxy so aqua-proxy is executed.
Then aqua-proxy executes `aqua exec` passing the program name and command line arguments.
In case of `go version`, `aqua exec -- go version` is executed.
`aqua exec` reads the configuration file and finds the package which includes `go` and gets the package versions.
aqua installs the package version in `~/.aqua/pkgs` if it isn't installed yet
Then aqua executes the command `~/.aqua/pkgs/http/golang.org/dl/go1.17.darwin-amd64.tar.gz/go/bin/go version`.

`~/.aqua/bin` is shared by every `aqua.yaml`, so maybe in `aqua exec` the package isn't found.
Please comment out the package `go` and execute `go version` again.

```yaml
inline_registry:
  packages:
  - name: go
    type: http
    url: https://golang.org/dl/go{{.Version}}.{{.OS}}-{{.Arch}}.tar.gz
    link: https://golang.org/
    description: The Go programming language
    files:
    - name: go # the symbolic `go` is created
      src: go/bin/go
    - name: gofmt # the symbolic `gofmt` is created
      src: go/bin/gofmt

# packages:
# - name: go
#   registry: inline
#   version: "1.17"
```

No package which includes `go` is found, so aqua checks the global configuration `~/.aqua/global/aqua.yaml`.
If the package isn't found in the global configuration too,
aqua finds the command from the environment variable `PATH`.
For example, if the `PATH` is `~/.aqua/bin:/usr/local/bin:/bin`, aqua checks the following files.

1. `~/.aqua/bin/go`
1. `/usr/local/bin/go`
1. `/bin/go`

To prevent the infinite loop, aqua ignores the symbolic to aqua-proxy.
`~/.aqua/bin/go` is a symbolic link to aqua-proxy, so this is ignored.
If go is installed in `/usr/local/bin/go`, `/usr/local/bin/go version` is executed.
If `go` isn't found, aqua exits with non zero exit code.
