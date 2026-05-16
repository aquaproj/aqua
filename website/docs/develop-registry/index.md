# Develop a Registry

You can install tools registered in the [Standard Registry](/docs/products/aqua-registry) easily.
You can search packages from Registries by `aqua g`.
Please see [Search Packages](/docs/tutorial/search-packages).

If tools aren't found, please [send a pull request to Standard Registry](/docs/products/aqua-registry/contributing).
It is okay only to [create an Issue](https://github.com/aquaproj/aqua-registry/issues) if it is difficult to send a pull request.

If tools are not private Repositories, we recommend sending a pull request to Standard Registry rather than maintaining them in your custom Registries, because

- You can get the support from maintainers
- You don't have to maintain Registry yourself
- [From aqua v2, you have to allow non Standard Registry explicitly for security](/docs/reference/upgrade-guide/v2/only-standard-registry-is-allowed-by-default/), but this is a bit bothersome

If tools are hosted in private repositories, please create custom Registries.

## Scaffold Registry Configuration

You can scaffold Registry Configuration by `aqua gr` command.

e.g.

```sh
aqua gr suzuki-shunsuke/ghalint > registry.yaml
```

By default, `aqua gr` generates Registry Configuration supporting all versions including old versions.
But if you only have to support only the latest version, you can set the option `-l 1`.

```sh
aqua gr -l 1 suzuki-shunsuke/ghalint
```

`aqua gr` command is imperfect, so sometimes you have to modify generated configuration yourself, but it's much easier than writing configuration from scratch.

If the command name is different from the package's repository name, you should set `-cmd` option.

e.g.

```sh
aqua gr -cmd gh cli/cli
```

You can use generated configuration as a local Registry or a github_content Registry.

## Allow private Registries by Policy

By default, aqua allows us to use only [Standard Registry](https://github.com/aquaproj/aqua-registry) for security.
To use private Registries, you have to allow them by [Policy](/docs/guides/policy-as-code).

## Use as a local Registry

aqua.yaml

```yaml
registries:
  - name: foo
    type: local
    path: registry.yaml # Relative path from aqua.yaml

packages:
  - name: suzuki-shunsuke/tfcmt@v3.2.4
    registry: foo
```

## Use as a github_content Registry

Add a Registry file to a GitHub Repository and push a tag for versioning.
Then you can use it as a github_content Registry.

aqua.yaml

```yaml
registries:
  - name: foo
    type: github_content
    repo_owner: suzuki-shunsuke
    repo_name: private-aqua-registry
    ref: v0.1.0
    path: registry.yaml

packages:
  - name: suzuki-shunsuke/tfcmt@v3.2.4
    registry: foo
```

If the Registry is private, you have to set a GitHub Access Token to the environment variable `AQUA_GITHUB_TOKEN` or `GITHUB_TOKEN`.
