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

There are several approaches to share aqua configurations across your team:

### Option 1: Git Repository (Recommended)

Create a repository in your GitHub Organization and add aqua configuration files for your teams and organization.

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

### Option 2: HTTP Registry (aqua >= v2.56.0)

For organizations with internal infrastructure, you can host registries on HTTP servers. This is useful when you want to:
- Avoid GitHub dependencies
- Use existing internal artifact repositories
- Have centralized control over tool versions

Example configuration:

```yaml
# aqua.yaml
registries:
  - name: company-tools
    type: http
    url: https://tools.company.com/aqua/{{.Version}}/registry.yaml
    version: v1.2.0

packages:
  - name: company/internal-tool
    registry: company-tools
    version: v3.0.0
```

The HTTP registry can be hosted on any HTTP server (nginx, Apache, S3 with static hosting, etc.) and supports both raw YAML files and tar.gz archives. See [HTTP Registry documentation](/docs/reference/config/#http-registry) for more details.
