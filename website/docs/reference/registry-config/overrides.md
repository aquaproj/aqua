---
sidebar_position: 1600
---

# `overrides`

aqua >= v1.3.0

[#607](https://github.com/aquaproj/aqua/issues/607)

You can override the following attributes on the specific `GOOS` and `GOARCH`.

- asset
- checksum
- complete_windows_ext
- files
- format
- replacements
- type
- url
- windows_ext

e.g. On Linux ARM64, `Arch` becomes `aarch64`.

```yaml
  overrides:
  - goos: linux
    replacements:
      arm64: aarch64
```

In case of `replacements`, maps are merged.

`goos` or `goarch` or `envs` is required.

e.g.

```yaml
  asset: arkade
  overrides:
  - goos: linux
    goarch: arm64
    asset: 'arkade-{{.Arch}}'
  - goos: darwin
    goarch: amd64
    asset: 'arkade-darwin'
  - goos: darwin 
    asset: 'arkade-darwin-{{.Arch}}'
```

Even if multiple elements are matched, only first element is applied.
For example, Darwin AMD64 matches with second element but the second element isn't applied because the first element is matched.

```yaml
  - goos: darwin
    goarch: amd64
    asset: 'arkade-darwin'
  - goos: darwin 
    asset: 'arkade-darwin-{{.Arch}}'
```

## envs

[#2318](https://github.com/aquaproj/aqua/issues/2318) [#2320](https://github.com/aquaproj/aqua/pull/2320) aqua >= [v2.13.0](https://github.com/aquaproj/aqua/releases/tag/v2.13.0)

You can use `envs` instead of `goos` and `goarch`.
The syntax of `envs` is same with [supported_envs](supported-envs.md).
`envs` is more flexible than the combination of `goos` and `goarch`, so in some cases you can simplify the code.

e.g.

`goos` and `goarch`

```yaml
overrides:
  - goos: windows
    goarch: arm64
    # ...
  - goos: linux
    goarch: arm64
    # ...
```

`envs` can simplify the code.

```yaml
overrides:
  - envs:
      - windows/arm64
      - linux/arm64
    # ...
```
