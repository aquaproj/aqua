---
sidebar_position: 20
---

# Share aqua configuration for teams and organizations

aqua reads configuration from the environment variable `AQUA_GLOBAL_CONFIG`.
`AQUA_GLOBAL_CONFIG` is configuration file paths separated with semicolons `:`.

e.g.

```sh
export AQUA_GLOBAL_CONFIG="/home/foo/aqua-config/sre.yaml:/home/foo/aqua-config/all.yaml:${AQUA_GLOBAL_CONFIG:-}"
```

About the priority of configuration, please see [Configuration File Path](/docs/reference/config/#configuration-file-path).

By default `aqua i` ignores Global Configuration.
If `--all (-a)` is set, aqua installs all packages including Global Configuration.

## How to share aqua configuration for teams and organizations

Let's create the repository in your GitHub Organization,
and add aqua configuration files for your teams and organization into the repository.

e.g.

```
aqua-config/
  all.yaml # aqua configuration for all developers in your organization
  sre.yaml # aqua configuration for your SRE team
```

Then checkout the repository and set the environment variable `AQUA_GLOBAL_CONFIG`.
If you belong to SRE team,

```sh
export AQUA_GLOBAL_CONFIG="/home/foo/aqua-config/sre.yaml:/home/foo/aqua-config/all.yaml:${AQUA_GLOBAL_CONFIG:-}"
```

Otherwise

```sh
export AQUA_GLOBAL_CONFIG="/home/foo/aqua-config/all.yaml:${AQUA_GLOBAL_CONFIG:-}"
```
