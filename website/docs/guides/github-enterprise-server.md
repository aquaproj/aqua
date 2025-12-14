---
sidebar_position: 405
---

# GitHub Enterprise Server (GHES) Support

`aqua >= v2.56.0` [#4361](https://github.com/aquaproj/aqua/pull/4361)

aqua supports using registries and packages hosted on GitHub Enterprise Server instances. This allows you to use aqua in enterprise environments where tools are hosted on internal GHES instances.

## Feature Overview

With GHES support, you can:

- Use **registries** hosted on GitHub Enterprise Server
- Authenticate with GHES using environment variables
- Use Policy as Code to control GHES registry access

:::info
**Important Limitation**: Currently, only registries hosted on GHES are supported. Direct package downloads from GHES repositories (e.g., `github_release` type packages on GHES) are **not yet supported**.

For example, if your GHES registry defines a `github_release` package, that package must still be released on github.com, not on your GHES instance.
:::

## Configuration

### Registry Configuration

To use a registry hosted on GHES, add the `github_enterprise_base_url` field to your registry configuration:

```yaml
registries:
  - name: ghes-registry
    type: github_content
    repo_owner: my-org
    repo_name: aqua-registry
    github_enterprise_base_url: https://ghes.example.com
    ref: v1.0.0
    path: registry.yaml
    private: true

packages:
  - name: my-org/my-tool@v1.0.0
    registry: ghes-registry
```

### Field Reference

- `github_enterprise_base_url`: The base URL of your GHES instance (e.g., `https://ghes.example.com`)
  - **Required** for GHES registries
  - Must be a valid HTTPS URL
  - Do not include the `/api/v3/` suffix (it will be added automatically)
- All other registry fields work the same as standard `github_content` registries

## Authentication

aqua supports authentication for GHES using environment variables. The authentication token is selected in the following priority order:

1. Domain-specific environment variables
2. Fallback GHES token

### Environment Variables

#### Domain-Specific Tokens (Recommended)

Set a token specific to your GHES domain:

```sh
export AQUA_GITHUB_TOKEN_ghes_example_com="your-ghes-token"
# or
export GITHUB_TOKEN_ghes_example_com="your-ghes-token"
```

**Important**: Replace dots (`.`) in the domain with underscores (`_`).

Examples of domain transformation:
- `ghes.example.com` → `GITHUB_TOKEN_ghes_example_com`
- `github.internal.company.com` → `GITHUB_TOKEN_github_internal_company_com`
- `ghe-server.local` → `GITHUB_TOKEN_ghe-server_local`

#### Fallback Token

Use `GITHUB_ENTERPRISE_TOKEN` as a fallback for any GHES instance when a domain-specific token is not set:

```sh
export GITHUB_ENTERPRISE_TOKEN="your-ghes-token"
```

This is useful when you have multiple GHES instances and want to use the same token for all of them.

## Token Priority

When accessing a GHES registry, aqua looks for tokens in this order:

1. `AQUA_GITHUB_TOKEN_<domain>` (e.g., `AQUA_GITHUB_TOKEN_ghes_example_com`)
2. `GITHUB_TOKEN_<domain>` (e.g., `GITHUB_TOKEN_ghes_example_com`)
3. `GITHUB_ENTERPRISE_TOKEN` (fallback for any GHES)

Note: For github.com, aqua uses `AQUA_GITHUB_TOKEN` or `GITHUB_TOKEN` as usual.

## Policy as Code

To use GHES registries, you need to allow them in your [Policy](/docs/reference/security/policy-as-code) file.

### Policy Configuration

Create or update your `aqua-policy.yaml`:

```yaml
registries:
  - type: standard
    ref: semver(">= 3.0.0")
  - name: ghes-registry
    type: github_content
    repo_owner: my-org
    repo_name: aqua-registry
    github_enterprise_base_url: https://ghes.example.com
    ref: semver(">= 1.0.0")
    path: registry.yaml

packages:
  - registry: standard
  - registry: ghes-registry  # Allow all packages from GHES registry
```

The `github_enterprise_base_url` field in the policy must match the value in your `aqua.yaml` registry configuration.

:::info
After updating your policy, run `aqua policy allow` to reflect changes. See [Policy as Code](/docs/guides/policy-as-code).
:::

## Requirements

- aqua v2.56.0 or later
- GitHub Enterprise Server instance with API access
- Personal Access Token or GitHub App token with appropriate permissions:
  - `repo` scope for private repositories
  - `contents:read` permission for reading repository contents

## Related Documentation

- [Configuration Reference - GHES Registry](/docs/reference/config/#github-enterprise-server-ghes-registry)
- [Develop a Registry - GHES Registry](/docs/develop-registry/#github-enterprise-server-ghes-registry)
- [Install Private Packages](/docs/guides/private-package)
- [Policy as Code](/docs/guides/policy-as-code)
