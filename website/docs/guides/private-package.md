---
sidebar_position: 400
---

# Install private packages

Of course, [Standard Registry](https://github.com/aquaproj/aqua-registry) doesn't include private packages, which are hosted in private GitHub Repositories.
But you can install private packages by creating your own custom Registry.

aqua supports several registry types for private packages:

- **[Local Registry](/docs/reference/config/#local-registry)**: Use a local file
- **[GitHub Content Registry](/docs/reference/config/#github_content-registry)**: Host your registry in a private GitHub repository
- **[HTTP Registry](/docs/reference/config/#http-registry)** (aqua >= v2.56.0): Host your registry on an internal HTTP server

The HTTP registry type is particularly useful for teams that want to host their registry on internal infrastructure without requiring GitHub:

```yaml
registries:
  - name: my-company
    type: http
    url: https://internal.company.com/aqua/{{.Version}}/registry.yaml
    version: v1.0.0
    format: raw

packages:
  - name: company/internal-tool
    registry: my-company
    version: v2.0.0
```

About how to create a Registry, please see [Develop a Registry](/docs/develop-registry/).
