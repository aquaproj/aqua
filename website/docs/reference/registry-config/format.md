---
sidebar_position: 610
---

# format

The asset format (e.g. `zip`, `tar.gz`).

e.g.

```yaml
format: tar.gz
```

This is used to unarchive or decompress the asset.
If this isn't specified, aqua tries to specify the format from the file extension.
If the file isn't archived and isn't compressed, please specify `raw`

```yaml
format: raw
```

## Support short file extensions

[#1876](https://github.com/aquaproj/aqua/issues/1876) [#2313](https://github.com/aquaproj/aqua/pull/2313) [v2.13.0](https://github.com/aquaproj/aqua/releases/tag/v2.13.0)

As of aqua v2.13.0, the following short file extensions are also supported.

- tbr
- tbz
- tbz2
- tgz
- tlz4
- tsz
- txz
