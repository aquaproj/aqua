---
sidebar_position: 1250
---

# Minisign

- `aqua >= v2.31.0` [#2978](https://github.com/aquaproj/aqua/pull/2978) [#2994](https://github.com/aquaproj/aqua/pull/2994)
- `aqua >= v2.34.0` [#3103](https://github.com/aquaproj/aqua/pull/3103)

aqua supports verifying packages with [minisign](https://github.com/jedisct1/minisign) to install some packages securely.
For example, [zig](https://ziglang.org/download/) is signed by minisign.

## Example

```yaml
packages:
  - type: http
    repo_owner: ziglang
    repo_name: zig
    # ...
    minisign:
      type: http
      url: https://ziglang.org/builds/zig-{{.OS}}-{{.Arch}}-{{.Version}}.{{.Format}}.minisig
      public_key: "RWSGOq2NVecA2UPNdBUZykf1CCb147pkmdtYxgb3Ti+JO/wCYvhbAb/U"
```

Verifying checksum files using Minisign.

```yaml
packages:
  - type: github_release
    repo_owner: bufbuild
    repo_name: buf
    asset: buf-{{.OS}}-{{.Arch}}.{{.Format}}
    format: tar.gz
    files:
      - name: buf
        src: buf/bin/buf
    replacements:
      amd64: x86_64
      darwin: Darwin
      linux: Linux
      windows: Windows
    checksum:
      type: github_release
      asset: sha256.txt
      algorithm: sha256
      minisign: # Minisign
        type: github_release
        asset: sha256.txt.minisig
        public_key: RWQ/i9xseZwBVE7pEniCNjlNOeeyp4BQgdZDLQcAohxEAH5Uj5DEKjv6
    overrides:
      - goos: linux
        replacements:
          arm64: aarch64
```
