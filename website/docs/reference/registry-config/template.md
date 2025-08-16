---
sidebar_position: 1700
---

# Template String

Some fields are parsed with [Go's text/template](https://pkg.go.dev/text/template) and [sprig](http://masterminds.github.io/sprig/).

:::caution
The following sprig functions are removed for security reason.

* [env](http://masterminds.github.io/sprig/os.html)
* [expandenv](http://masterminds.github.io/sprig/os.html)
* [getHostByName](http://masterminds.github.io/sprig/network.html)
:::

## Common Template Functions

* `trimV`: This is equivalent to `trimPrefix "v"`

## Template Variables

* `OS`: A string which `GOOS` is replaced by `replacements`. If `replacements` isn't set, `OS` is equal to `GOOS`. Basically you should use `OS` for the consistency
* `Arch`: A string which `GOARCH` is replaced by `replacements`. If `replacements` isn't set, `Arch` is equal to `GOARCH`. Basically you should use `OS` for the consistency
* `GOOS`: Go's [runtime.GOOS](https://pkg.go.dev/runtime#pkg-constants)
* `GOARCH`: Go's [runtime.GOARCH](https://pkg.go.dev/runtime#pkg-constants)
* `Version`: Package `version`
* `SemVer`: Package version that [version_prefix](version-prefix.md) is trimmed from `Version`. For example, if `Version` is `cli/v1.0.0` and `version_prefix` is `cli/`, then `SemVer` is `v1.0.0`
* `Format`: Package `format`
* `FileName`: `files[].name`
* `AssetWithoutExt`

### AssetWithoutExt

[#1774](https://github.com/aquaproj/aqua/issues/1774) [#2310](https://github.com/aquaproj/aqua/pull/2310) aqua >= [v2.13.0](https://github.com/aquaproj/aqua/releases/tag/v2.13.0)

`AssetWithoutExt` is a string that a file extension is removed from `Asset`.

e.g.

```yaml
    asset: aks-engine-{{.Version}}-{{.OS}}-{{.Arch}}.tar.gz
    files:
      - name: aks-engine
        src: "{{.AssetWithoutExt}}/aks-engine" # "{{.AssetWithoutExt}}" == "aks-engine-{{.Version}}-{{.OS}}-{{.Arch}}"
```

### Omit `Format` in `asset` and `url`

[#1774](https://github.com/aquaproj/aqua/issues/1774) [#2314](https://github.com/aquaproj/aqua/pull/2314) aqua >= [v2.13.0](https://github.com/aquaproj/aqua/releases/tag/v2.13.0)

The file extension is complemented if it isn't included in `asset` and `url`.

e.g.

```yaml
asset: foo # This is same with `foo.tar.gz` and `foo.{{.Format}}`
format: tar.gz
```

You can unify the setting of `raw` format and non `raw` format.

Before

```yaml
asset: kalker-{{.OS}}.{{.Format}}
format: zip
overrides:
  - goos: linux
    format: raw
    asset: kalker-{{.OS}}
```

After

```yaml
asset: kalker-{{.OS}}
format: zip
overrides:
  - goos: linux
    format: raw
```

You can disable the complementation by setting `append_ext: false`.

```yaml
append_ext: false
```

By default `append_ext` is `true`.
