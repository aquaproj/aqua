---
sidebar_position: 1800
---

# `version_constraint`, `version_overrides`

Some package attributes like `asset` and `files` may be different by it's version.
For example, the asset structure of [golang-migrate](https://github.com/golang-migrate/migrate) was changed from v4.15.0.
In that case, the attributes `version_constraint` and `version_overrides` are useful.

e.g.

```yaml
- type: github_release
  repo_owner: golang-migrate
  repo_name: migrate
  asset: 'migrate.{{.OS}}-{{.Arch}}.tar.gz'
  description: Database migrations. CLI and Golang library
  version_constraint: 'semver("> 4.14.1")'
  version_overrides:
  - version_constraint: 'semver("<= 4.14.1")'
    files:
    - name: migrate
      src: 'migrate.{{.OS}}-{{.Arch}}'
```

`version_constraint` is [expr](https://github.com/antonmedv/expr)'s expression.
The evaluation result must be a boolean.

Currently, the following values and functions are accessible in the expression.

* `Version`: (type: `string`) The package version
* `SemVer`: (type: `string`) The package version that [version_prefix](version-prefix.md) is trimmed from `Version`. For example, if `Version` is `cli/v1.0.0` and `version_prefix` is `cli/`, then `SemVer` is `v1.0.0`
* `semver`: (type: `func(string) bool`) Tests if the package version satisfies all the constraints
* `semverWithVersion`: (type: `func(constraint, version string) bool`) Tests if the package version satisfies all the constraints
* `trimPrefix`: (type `func(s, prefix string) string`) Go's [strings.TrimPrefix](https://pkg.go.dev/strings#TrimPrefix)

### semver, semverWithVersion

[See also `Change the implementation of semver and semverWithVersion`](/docs/reference/upgrade-guide/v2/change-semver).

e.g.

```
semver("> 4.14.1")
```

```
semver("> 3.0.0, <= 4.0.0")
```

`semverWithVersion` is used when we need to format `Version`:

```
semverWithVersion(">= 0.11.1", trimPrefix(Version, "cli-"))
```

Syntax:

```
<operator> <version>[, <operator> <version> ...]
```

Supported operators:

- `>=`
- `<=`
- `!=`
- `>`
- `<`
- `=`

Support multiple constraints separated with comma `,`.
Multiple constraints are evaluated as `AND` comparison.
Spaces are trimmed.
Constraints are evaluated using [hashicorp/go-version](https://github.com/hashicorp/go-version).

## version_overrides

The list of version override.

e.g.

```yaml
  version_overrides:
  - version_constraint: 'semver("<= 4.14.1")'
    files:
    - name: migrate
      src: 'migrate.{{.OS}}-{{.Arch}}'
```
