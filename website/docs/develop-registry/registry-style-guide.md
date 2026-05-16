---
sidebar_position: 445
---

# Registry Style Guide

:::caution
If you develop your custom registry, you don't have to conform to this style guide.
But the guide would be useful to write good and consistent configuration.
:::

## Format with prettier

https://prettier.io/

```console
$ npm i -g prettier
```

```console
$ prettier -w registry.yaml
```

## Remove spaces in the template `{{ ` and ` }}`

:thumbsup:

```yaml
asset: tfcmt_{{.OS}}_{{.Arch}}.tar.gz
```

:thumbsdown:

```yaml
asset: tfcmt_{{ .OS }}_{{ .Arch }}.tar.gz
```

## Remove characters `.!` from the end of the description

:thumbsup:

```yaml
description: A command-line tool that makes git easier to use with GitHub
```

:thumbsdown:

```yaml
description: A command-line tool that makes git easier to use with GitHub.
```

## Trim spaces

:thumbsup:

```yaml
description: A command-line tool that makes git easier to use with GitHub
```

:thumbsdown:

```yaml
description: "  A command-line tool that makes git easier to use with GitHub  "
```

## Remove unneeded quotes of strings

:thumbsup:

```yaml
description: A command-line tool that makes git easier to use with GitHub
```

:thumbsdown:

```yaml
description: "A command-line tool that makes git easier to use with GitHub"
```

## Avoid `if` and `for` statement in templates

:thumbsup:

```yaml
asset: foo.{{.Format}}
format: tar.gz
overrides:
  - goos: windows
    format: zip
```

:thumbsdown:

```yaml
asset: 'foo.{{if eq .GOOS "windows"}}zip{{else}}tar.gz{{end}}'
```

## `version_overrides` Style Guide

We decided not to rely on base settings as much as possible.
This means we don't define settings such as `asset`, `format`, `windows_arm_emulation`, and so on on the base settings.
Merge with base settings makes code DRY, but it's difficult to update settings when settings of new versions are changed because the update of the base settings affects all version override.
By stopping to merge settings, we can update settings by simply adding a new version override and updating the last version_constraint.
Perhaps we would be able to automate the update in future too.

e.g.

```yaml
# Define only settings which don't depend on versions.
# e.g. repo_owner, repo_name, description.
version_constraint: "false"
version_overrides:
  - version_constraint: semver("<= 3.0.0")
    # Oldest setting
    # ...
  - version_constraint: semver("<= 4.0.0")
    # ...
  - version_constraint: semver("<= 5.0.0")
    # ...
  - version_constraint: "true"
    # Latest setting
    # ...
```

## If the `format` is `raw`, `files[].src` isn't needed

:thumbsup:

```yaml
format: raw
files:
  - name: swagger
```

:thumbsdown:

```yaml
format: raw
files:
  - name: swagger
    src: swagger_{{.OS}}_{{.Arch}} # unneeded
```

## musl vs glibc

When a package provides pre-built binaries for both musl and glibc, which one to install depends on whether the musl build is static or dynamic (depends on libc).
You can use the `file` command to determine whether it is static.
If the `file` command is not available, you can use `ldd`.
You can also use `readelf -d` or `objdump -p | grep NEEDED` as safer alternatives that don't execute the binary — if there is a `NEEDED` entry, it is dynamic; otherwise static.

For example, the musl build of claude code v2.1.126 is dynamically linked.

```console
50cbd103ea9b:/workspace# file "$(aqua which claude)"
/root/.local/share/aquaproj-aqua/pkgs/http/storage.googleapis.com/claude-code-dist-86c565f3-f756-42ad-8dfa-d59b1c096819/claude-code-releases/2.1.126/linux-arm64-musl/claude/claude: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), dynamically linked, interpreter /lib/ld-musl-aarch64.so.1, BuildID[sha1]=0a41f4e4ceddb1b805d2f5c94fdffd91bf366fb4, not stripped
```

```console
50cbd103ea9b:/workspace# ldd "$(aqua which claude)"
	/lib/ld-musl-aarch64.so.1 (0xf1e6bb2c0000)
	libc.musl-aarch64.so.1 => /lib/ld-musl-aarch64.so.1 (0xf1e6bb2c0000)
```

```console
50cbd103ea9b:/workspace# readelf -d "$(aqua which claude)"

Dynamic section at offset 0x6100000 contains 25 entries:
  Tag        Type                         Name/Value
 0x0000000000000001 (NEEDED)             Shared library: [libc.musl-aarch64.so.1]
 ...
```

```console
50cbd103ea9b:/workspace# objdump -p "$(aqua which claude)" | grep NEEDED
  NEEDED               libc.musl-aarch64.so.1
```

1. If the musl build is static: always use the musl build
1. If the musl build is dynamic: use the glibc build by default, and use the musl build via [variants](/docs/reference/registry-config/overrides#variants) in environments where the musl build works

### 1. When the musl build is static

While installing the glibc build is an option in environments where it works, we adopt the musl build for several reasons.

1. The musl build works fine
2. Always installing the musl build is simpler
3. The glibc build may not work depending on the glibc version
4. Currently aqua has no feature to check the glibc version. Although it is possible to implement, there is little benefit, so we have no plans to do so
5. The benefits of the glibc build (such as file size) are not significant, and the benefits of always installing the musl build outweigh them

### 2. When the musl build is dynamic

Use the glibc build by default, and use the musl build via [variants](/docs/reference/registry-config/overrides#variants) in environments where the musl build works.
Using the glibc build by default matches the behavior from before variants were supported, preserving compatibility.
However, in environments such as Alpine where the glibc build does not work but the musl build does, we use the musl build (this requires aqua >= v2.58.0, which supports variants).

Please see also [variants](/docs/reference/registry-config/overrides#variants).

## Consideration about Rust

:warning: The author [@suzuki-shunsuke](https://github.com/suzuki-shunsuke) isn't familiar with Rust. If you have any opinion, please let us know.

- linux: use the asset for not `glibc` but `musl` if both of them are supported
  - ref: https://github.com/aquaproj/aqua-registry/pull/2153#discussion_r805116879
- windows: use the asset for not `gnu` but `msvc` if both of them are supported
  - ref: https://rust-lang.github.io/rustup/installation/windows.html

:thumbsup:

```yaml
replacements:
  linux: unknown-linux-musl
  windows: pc-windows-msvc
```

:thumbsdown:

```yaml
replacements:
  linux: unknown-linux-gnu
  windows: pc-windows-gnu
```

## Use `overrides` instead of `format_overrides`

:thumbsup:

```yaml
format: tar.gz
overrides:
  - goos: windows
    format: zip
```

:thumbsdown:

```yaml
format: tar.gz
format_overrides:
  - goos: windows
    format: zip
```

## Don't use emojis as much as possible

In some environments, emojis are corrupted. e.g. https://github.com/aquaproj/aqua/pull/1004#issuecomment-1183710603

:thumbsup:

```yaml
description: CLI and Go library for CODEOWNERS files
```

:thumbsdown:

```yaml
description: 🔒 CLI and Go library for CODEOWNERS files
```

## Omit the setting which is equivalent to the default value

When `repo_owner` and `repo_name` are set, you can omit some attributes.

:thumbsup:

```yaml
repo_owner: weaveworks
repo_name: eksctl
```

:thumbsdown:

```yaml
repo_owner: weaveworks
repo_name: eksctl
name: weaveworks/eksctl
link: https://github.com/weaveworks/eksctl
files:
  - name: eksctl
```

## Use `aliases` only for keeping the compatibility

Please see [here](/docs/reference/registry-config/aliases#use-aliases-only-for-keeping-the-compatibility)

## Use `supported_envs` rather than `supported_if`

Please see [the caution](/docs/reference/registry-config/supported-if).

## Select `type` according to the following order

1. github_release
1. github_content
1. github_archive
1. http
1. go_install
1. go_build

For example, you can also use `http` type to install the package from GitHub Releases, but in that case you should use `github_release` rather than `http`.

## `cargo` package name should be `crates.io/<crate name>`

Please see [here](/docs/reference/registry-config/cargo-package#-package-name).
