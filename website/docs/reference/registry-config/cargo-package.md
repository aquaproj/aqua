---
sidebar_position: 1310
---

# `cargo` Package

[#2016](https://github.com/aquaproj/aqua/discussions/2016) [#2019](https://github.com/aquaproj/aqua/pull/2019), aqua >= [v2.8.0](https://github.com/aquaproj/aqua/releases/tag/v2.8.0)

The package is installed by [cargo install](https://doc.rust-lang.org/cargo/commands/cargo-install.html) command.

You can manage tools written in Rust with aqua, which means you can manage them and their versions declaratively in the consistent way. You can switch tool versions per project and update them continuously by Renovate!

:::info
We aren't familiar with Rust and Cargo, so your contribution is welcome!
:::

## Requirements

Please install these tools in advance.

- [Rust](https://www.rust-lang.org/)
- [Cargo](https://doc.rust-lang.org/cargo/)

## Fields

- `crate` (required): Crate name. e.g. [skim](https://crates.io/crates/skim)
- `cargo`: `cargo install` command options
  - `all_features` (boolean): `--all-features` option
  - `features` ([]string): `--features` option
  - `locked` (aqua >= v2.41.0): `--locked` option

## Try `cargo` package quickly

```console
$ aqua init
$ aqua g -i crates.io/skim
$ aqua i -l
$ sk -V
```

## Example

registry.yaml

```yaml
packages:
  - name: crates.io/skim
    type: cargo
    crate: skim
    files:
      - name: sk
  - name: crates.io/broot
    type: cargo
    crate: broot
    cargo:
      locked: true
      all_features: true
      # features:
      #   - clipboard
```

aqua.yaml

```yaml
registries:
  - name: local
    type: local
    path: registry.yaml
packages:
  - name: crates.io/skim@0.10.4
    registry: local
```

### aqua i

```console
$ aqua i
```

:::info
cargo outputs the following warning, but there is no problem. Please ignore it.

```
warning: be sure to add `/home/foo/.local/share/aquaproj-aqua/pkgs/cargo/crates.io/skim/0.10.4/bin` to your PATH to be able to run the installed binaries
```
:::

:::caution
If `cargo install` fails, please see [here](/docs/reference/codes/005).
:::

<details>
<summary>$ aqua i</summary>

```console
$ aqua i
INFO[0000] Installing a crate                            aqua_version= env=darwin/arm64 package_name=crates.io/skim package_version=0.10.4 program=aqua registry=local
    Updating crates.io index
  Installing skim v0.10.4
   Compiling autocfg v1.1.0
   Compiling cfg-if v1.0.0
   Compiling libc v0.2.144
   Compiling proc-macro2 v1.0.58
   Compiling unicode-ident v1.0.8
   Compiling quote v1.0.27
   Compiling crossbeam-utils v0.8.15
   Compiling syn v1.0.109
   Compiling fnv v1.0.7
   Compiling strsim v0.10.0
   Compiling ident_case v1.0.1
   Compiling memchr v2.5.0
   Compiling memoffset v0.8.0
   Compiling crossbeam-epoch v0.9.14
   Compiling num-traits v0.2.15
   Compiling scopeguard v1.1.0
   Compiling crossbeam-channel v0.5.8
   Compiling num-integer v0.1.45
   Compiling log v0.4.17
   Compiling once_cell v1.17.1
   Compiling bitflags v1.3.2
   Compiling memoffset v0.6.5
   Compiling indexmap v1.9.3
   Compiling dirs-sys-next v0.1.2
   Compiling crossbeam-deque v0.8.3
   Compiling rayon-core v1.11.0
   Compiling core-foundation-sys v0.8.4
   Compiling crossbeam-queue v0.3.8
   Compiling iana-time-zone v0.1.56
   Compiling aho-corasick v1.0.1
   Compiling dirs-next v2.0.0
   Compiling atty v0.2.14
   Compiling num_cpus v1.15.0
   Compiling time v0.1.45
   Compiling regex-syntax v0.7.1
   Compiling os_str_bytes v6.5.0
   Compiling hashbrown v0.12.3
   Compiling termcolor v1.2.0
   Compiling clap_lex v0.2.4
   Compiling chrono v0.4.24
   Compiling darling_core v0.14.4
   Compiling term v0.7.0
   Compiling nix v0.24.3
   Compiling regex v1.8.1
   Compiling thread_local v1.1.7
   Compiling vte_generate_state_changes v0.1.1
   Compiling textwrap v0.16.0
   Compiling lazy_static v1.4.0
   Compiling either v1.8.1
   Compiling pin-utils v0.1.0
   Compiling arrayvec v0.7.2
   Compiling unicode-width v0.1.10
   Compiling utf8parse v0.2.1
   Compiling humantime v2.1.0
   Compiling darling_macro v0.14.4
   Compiling time-core v0.1.1
   Compiling vte v0.11.0
   Compiling env_logger v0.9.3
   Compiling time v0.3.21
   Compiling tuikit v0.5.0
   Compiling darling v0.14.4
   Compiling clap v3.2.25
   Compiling derive_builder_core v0.11.2
   Compiling nix v0.25.1
   Compiling rayon v1.7.0
   Compiling fuzzy-matcher v0.3.7
   Compiling timer v0.2.0
   Compiling derive_builder_macro v0.11.2
   Compiling derive_builder v0.11.2
   Compiling crossbeam v0.8.2
   Compiling defer-drop v1.3.0
   Compiling shlex v1.1.0
   Compiling beef v0.5.2
   Compiling skim v0.10.4
    Finished release [optimized] target(s) in 32.46s
  Installing /home/foo/.local/share/aquaproj-aqua/pkgs/cargo/crates.io/skim/0.10.4/bin/sk
   Installed package `skim v0.10.4` (executable `sk`)
warning: be sure to add `/home/foo/.local/share/aquaproj-aqua/pkgs/cargo/crates.io/skim/0.10.4/bin` to your PATH to be able to run the installed binaries
```

</details>

### aqua g

aqua gets versions by crates.io API.

```console
$ aqua g local,crates.io/skim
- name: crates.io/skim@0.10.4
  registry: local
```

### aqua gr

If the package name starts with `crates.io/`, `aqua gr` command treats the package as `cargo` type package.

```console
$ aqua gr crates.io/skim
packages:
  - name: crates.io/skim
    type: cargo
    repo_owner: lotabout
    repo_name: skim
    description: Fuzzy Finder in rust!
    crate: skim
```

:::info
In case of the package type `cargo`, you don't have to specify `--deep` option.
:::

You have to set `files` if necessary. This is same with other package types.

```yaml
packages:
  - name: crates.io/skim
    type: cargo
    repo_owner: lotabout
    repo_name: skim
    description: Fuzzy Finder in rust!
    crate: skim
    files: # Set files manually
      - name: sk
```

### :bulb: Package name

:::tip
If you add a crate hosted at [crates.io](https://crates.io/), we recommend the package name is `crates.io/<crate name>` such as `crates.io/skim` because

1. `aqua gr` and `aqua-registry gr` command can treat the package as `cargo` package
1. [aqua-renovate-config](/docs/products/aqua-renovate-config) can update the package
:::
