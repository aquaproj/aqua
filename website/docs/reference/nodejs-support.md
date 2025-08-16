---
sidebar_position: 450
---

# Node.js Support

You can manage Node.js using aqua.

`aqua-registry >= v4.216.0`

- [#2996](https://github.com/aquaproj/aqua/issues/2996)
- [aqua-registry#26002](https://github.com/aquaproj/aqua-registry/pull/26002)

## Set up

1. Configure npm's `prefix` config and the environment variable `$PATH`
1. Update aqua-registry to v4.216.0 or later
1. Install the package `nodejs/node`

### 1. Configure npm's `prefix` config and the environment variable `$PATH`

You need to configure npm's `prefix` config and the environment variable `$PATH`.

About npm's `prefix` config, please see the following links.

- https://docs.npmjs.com/cli/v10/commands/npm-install#global
- https://docs.npmjs.com/cli/v10/configuring-npm/folders#prefix-configuration
- https://docs.npmjs.com/cli/v10/using-npm/config#prefix
- https://docs.npmjs.com/cli/v10/using-npm/config#npmrc-files
- https://docs.npmjs.com/cli/v10/using-npm/config#environment-variables

```sh
export NPM_CONFIG_PREFIX="${XDG_DATA_HOME:-$HOME/.local/share}/npm-global" # You can change the path freely
export PATH=$NPM_CONFIG_PREFIX/bin:$PATH
```

In case of Windows, `bin` directory is missing.

```sh
export PATH=$NPM_CONFIG_PREFIX:$PATH
```

### 2. Update aqua-registry to v4.216.0 or later

```sh
aqua up -r
```

```yaml
registries:
  - type: standard
    ref: v4.216.0 # renovate: depName=aquaproj/aqua-registry
```

### 3. Install the package `nodejs/node`

```sh
aqua g -i nodejs/node
aqua i -l
```

## Example

```sh
node -v
npm -v
```

Install tools such as typescript and [zx](https://github.com/google/zx) by `npm i -g`.

```sh
npm i -g zx
zx -v
```

npm's `prefix` config was configured, so tools installed by `npm i -g` are shared between multiple Node.js version.
This has both pros and cons.

Pros.

- You don't have to reinstall tools when you change Node.js versions

```sh
aqua up node@v20.16.0 # Change Node.js version
zx -v # zx is still available
```

Cons.

- Sharing tools between multiple Node.js versions may cause some trouble. Tools may not work with other Node.js versions. In this case, you would have to change npm's `prefix` config by Node.js version somehow or reinstall tools

We decided to accept cons for now.
If you face any trouble about this, please let us know.

## Reference

If you're interested in the discussion about Node.js support, please check [#2996](https://github.com/aquaproj/aqua/issues/2996).
