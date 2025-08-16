---
sidebar_position: 310
---

# go_version_path

[#3269](https://github.com/aquaproj/aqua/pull/3269) [v2.38.0](https://github.com/aquaproj/aqua/releases/tag/v2.38.0)

If this field is set, `aqua g` and `aqua up` commands get versions from [Go Module Proxy](https://proxy.golang.org/).

```yaml
packages:
  - name: _go/sigsum.org/sigsum-go#cmd/sigsum-key
    type: go_install
    path: sigsum.org/sigsum-go/cmd/sigsum-key
    go_version_path: sigsum.org/sigsum-go
```

https://go.dev/ref/mod#communicating-with-proxies

```console
$ curl "https://proxy.golang.org/sigsum.org/sigsum-go/@v/list" | sort
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   408  100   408    0     0   6362      0 --:--:-- --:--:-- --:--:--  6476
v0.0.1
v0.0.2
v0.0.3
v0.0.4
v0.0.5
v0.0.6
v0.0.7
v0.0.8
v0.0.9
v0.1.0
v0.1.1
v0.1.10
v0.1.11
v0.1.12
v0.1.13
v0.1.14
v0.1.15
v0.1.16
v0.1.17
v0.1.18
v0.1.19
v0.1.2
v0.1.20
v0.1.21
v0.1.22
v0.1.23
v0.1.24
v0.1.25
v0.1.3
v0.1.4
v0.1.5
v0.1.6
v0.1.7
v0.1.8
v0.1.9
v0.2.0
v0.3.0
v0.3.1
v0.3.2
v0.3.3
v0.3.4
v0.3.5
v0.4.0
v0.4.1
v0.5.0
v0.6.0
v0.6.1
v0.6.2
v0.7.0
v0.7.1
v0.7.2
v0.8.0
v0.8.1
v0.8.2
v0.9.0
v0.9.1
```
