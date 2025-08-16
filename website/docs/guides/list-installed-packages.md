---
sidebar_position: 5
---

# List installed packages

aqua >= v2.24.0 [#2709](https://github.com/orgs/aquaproj/discussions/2709) [#2733](https://github.com/aquaproj/aqua/pull/2733)

The command `aqua list -installed` outputs installed packages.

```console
$ aqua list -installed
rhysd/actionlint	v1.6.27	standard
suzuki-shunsuke/cmdx	v1.7.4	standard
sigstore/cosign	v1.13.2	standard
suzuki-shunsuke/ghalint	v0.2.9	standard
int128/ghcp	v1.13.2	standard
golangci/golangci-lint	v1.56.2	standard
goreleaser/goreleaser	v1.24.0	standard
reviewdog/reviewdog	v0.17.1	standard
```

By default, global configuration files are ignored.
To output packages in global configuration files too, please set the option `-all [-a]`.

```console
$ aqua list -a -installed
```
