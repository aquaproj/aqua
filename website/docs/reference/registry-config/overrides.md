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

## variants

[#4514](https://github.com/aquaproj/aqua/issues/4514) [#4733](https://github.com/aquaproj/aqua/pull/4733) aqua >= v2.58.0

`variants` is an additional matching condition on `overrides` that distinguishes artifacts sharing the same `goos` / `goarch` along another axis (currently `libc`).
This is useful for upstreams that publish multiple builds per platform — for example, separate musl and glibc builds for Linux.

`variants` is a list of `key` / `value` pairs.
An override matches only when **every** variant's `value` equals the corresponding value detected by aqua at runtime.
An override without `variants` keeps the current behavior.

```yaml
overrides:
  - goos: linux
    goarch: amd64
    variants:
      - key: libc
        value: musl
    asset: foo-{{.OS}}-{{.Arch}}-musl.{{.Format}}

  - goos: linux
    goarch: amd64
    variants:
      - key: libc
        value: gnu
    asset: foo-{{.OS}}-{{.Arch}}-gnu.{{.Format}}
```

### Supported keys

| key | values | runtime detection |
|-----|--------|-------------------|
| `libc` | `musl`, `gnu` | On Linux: presence of `/lib/ld-musl-*.so.1` or `/lib/libc.musl-*.so.1`, otherwise `ldd --version` output. On non-Linux: not detected (empty). |

You can override the detected libc with the `AQUA_LIBC` environment variable (e.g. `AQUA_LIBC=musl`).

### Backward compatibility (aqua < v2.58.0)

aqua versions before v2.58.0 do not recognize the `variants` field — the YAML parser silently drops it.
As a result, an override carrying `variants` is matched against the existing conditions only (`goos` / `goarch` / `envs`), and aqua's "first matching entry wins" rule then picks whatever entry comes first in the list.

To keep registries compatible with older aqua versions, **place the override that older aqua should pick first**.

#### Example: `anthropics/claude-code`

`anthropics/claude-code` originally defaults to glibc (the URL without the `-musl` suffix).
See [aqua-registry#47592](https://github.com/aquaproj/aqua-registry/pull/47592).

##### Wrong: breaks older aqua

```yaml
version_overrides:
  - version_constraint: "true"
    url: https://storage.googleapis.com/claude-code-dist-86c565f3-f756-42ad-8dfa-d59b1c096819/claude-code-releases/{{trimV .Version}}/{{.OS}}-{{.Arch}}/claude
    overrides:
      - goos: linux
        url: https://storage.googleapis.com/claude-code-dist-86c565f3-f756-42ad-8dfa-d59b1c096819/claude-code-releases/{{trimV .Version}}/{{.OS}}-{{.Arch}}-musl/claude
        variants:
          - key: libc
            value: musl
```

On aqua `< v2.58.0`, `variants` is ignored, so this override matches every Linux user and the **musl URL** is selected — including for glibc users, who then end up with a binary that may not run.

##### Correct: preserves older behavior

Put the gnu override **first** so older aqua matches it before reaching the musl entry.
Its `url` can be omitted because the top-level `url` is inherited when an override does not set its own.

```yaml
version_overrides:
  - version_constraint: "true"
    url: https://storage.googleapis.com/claude-code-dist-86c565f3-f756-42ad-8dfa-d59b1c096819/claude-code-releases/{{trimV .Version}}/{{.OS}}-{{.Arch}}/claude
    overrides:
      - goos: linux
        variants:
          - key: libc
            value: gnu
        # url omitted → inherits the top-level (gnu) URL
      - goos: linux
        url: https://storage.googleapis.com/claude-code-dist-86c565f3-f756-42ad-8dfa-d59b1c096819/claude-code-releases/{{trimV .Version}}/{{.OS}}-{{.Arch}}-musl/claude
        variants:
          - key: libc
            value: musl
```

#### What each user sees

| aqua version | glibc user | musl user |
|----|----|----|
| `< v2.58.0` (`variants` ignored) | gnu URL ✅ (unchanged) | gnu URL ⚠️ (unchanged — same as before v2.58.0; the binary may not run on musl) |
| `>= v2.58.0` | gnu URL ✅ | musl URL ✅ |

musl users who want the musl binary need to upgrade to aqua `>= v2.58.0`.
On older aqua they continue to get the glibc URL, which is the same behavior as before `variants` existed — so this is not a regression.

### Introducing a new variant key in a future aqua version

Suppose a future aqua version (say v2.60.0) introduces a new variant key such as `linkage`, and you want to use it on newer aqua while keeping a sensible default on older versions.
There are three behavior ranges to keep in mind:

- aqua `< v2.58.0`: `variants` is dropped entirely by the YAML parser; overrides are matched on `goos` / `goarch` / `envs` only.
- aqua `>= v2.58.0, < v2.60.0`: `variants` is parsed, but `linkage` is unknown, so any override referencing it is skipped (skip-on-unknown).
- aqua `>= v2.60.0`: `linkage` is detected and matched normally.

#### Real-world example: containerd

`containerd` distributes both dynamically-linked and statically-linked Linux binaries (see the [containerd v2.2.3 release notes](https://github.com/containerd/containerd/releases/tag/v2.2.3)):

> - `containerd-<VERSION>-<OS>-<ARCH>.tar.gz`: ✅Recommended. Dynamically linked with glibc 2.35 (Ubuntu 22.04).
> - `containerd-static-<VERSION>-<OS>-<ARCH>.tar.gz`: Statically linked. Expected to be used on Linux distributions that do not use glibc >= 2.35. Not position-independent.

The static binary is the safer default — it works in more environments.
With the configuration below, every aqua version downloads the static binary by default, while aqua `>= v2.60.0` switches to the dynamic binary on hosts that can run it.

```yaml
asset: containerd-static-{{trimV .Version}}-{{.OS}}-{{.Arch}}.{{.Format}}
overrides:
  # On aqua >= v2.58.0, < v2.60.0, both overrides below are skipped because
  # the unknown `linkage` key disqualifies them, so the top-level (static)
  # asset is used.
  - goos: linux
    # No `asset` here: when this override matches, the top-level (static)
    # asset is inherited.
    # On aqua < v2.58.0, variants is ignored and this entry matches first.
    # On aqua >= v2.58.0, < v2.60.0, the unknown `linkage` key skips this.
    # On aqua >= v2.60.0 with a static-only host (one that cannot run the
    # dynamic build), this matches.
    variants:
      - key: linkage
        value: static
  - goos: linux
    # On aqua < v2.58.0, the entry above matches first, so this is unreachable.
    # On aqua >= v2.58.0, < v2.60.0, the unknown `linkage` key skips this.
    # On aqua >= v2.60.0 with a dynamic-capable host, this matches.
    asset: containerd-{{trimV .Version}}-{{.OS}}-{{.Arch}}.{{.Format}}
    variants:
      - key: linkage
        value: dynamic
```

#### Pattern: top-level default as a forward-compatible fallback

The example relies on a general pattern: **place the safest, most compatible asset at the top level, and let any override using a not-yet-recognized variant key fall through to it**.
This works as long as the top-level asset is the safer choice for environments where the new variant key cannot be resolved.

### Matching rules recap

1. `goos` / `goarch` / `envs` must match (existing behavior).
2. Every entry in `variants` must match the runtime value for the same key.
3. Variants whose `key` is not supported by the running aqua make the entire override non-matching.
4. The first matching `overrides` entry wins, as with any other override.
