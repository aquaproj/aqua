# Share aqua configuration for teams and organizations

[#321](https://github.com/suzuki-shunsuke/aqua/issues/321) [#322](https://github.com/suzuki-shunsuke/aqua/pull/322)

aqua reads configuration from the environment variable `AQUA_GLOBAL_CONFIG`.
`AQUA_GLOBAL_CONFIG` is configuration file paths separated with semicolons `:`.

e.g.

```sh
export AQUA_GLOBAL_CONFIG=/home/foo/aqua-config/sre.yaml:/home/foo/aqua-config/all.yaml
```

The priority of configuration is the following.

1. configuration file specified by `-c` option or configuration file which found from the current directory to the root directory
1. `$AQUA_GLOBAL_CONFIG`
1. `$AQUA_ROOT_DIR/global/[.]aqua.y[a]ml` (`~/.aqua/global/[.]aqua.y[a]ml`)

The option `--all (-a)` is added to the command `aqua install`.
By default `aqua install` ignores `$AQUA_GLOBAL_CONFIG` and `$AQUA_ROOT_DIR/global/[.]aqua.y[a]ml` (`~/.aqua/global/[.]aqua.y[a]ml`).
If `--all (-a)` is set, aqua installs all packages including `$AQUA_GLOBAL_CONFIG` and `$AQUA_ROOT_DIR/global/[.]aqua.y[a]ml` (`~/.aqua/global/[.]aqua.y[a]ml`).

## How to share aqua configuration for teams and organizations

I'll introduce one idea to share aqua configuration for teams and organizations.

Let's create the repository `aqua-config` in your GitHub Organization, and add aqua configuration files for your teams and organization into the repository.

```
aqua-config/
  all.yaml # aqua configuration for all developers in your organization
  sre.yaml # aqua configuration for your SRE team
```

Then checkout the repository and set the environment variable `AQUA_GLOBAL_CONFIG`.
If you belong to SRE team,

```sh
export AQUA_GLOBAL_CONFIG=/home/foo/aqua-config/sre.yaml:/home/foo/aqua-config/all.yaml
```

Otherwise

```sh
export AQUA_GLOBAL_CONFIG=/home/foo/aqua-config/all.yaml
```
