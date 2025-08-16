---
sidebar_position: 500
---

# Change the implementation of `semver` and `semverWithVersion`

[#1572](https://github.com/aquaproj/aqua/pull/1572) [#1573](https://github.com/aquaproj/aqua/issues/1573) 

The specification of `semver` and `semverWithVersion` are a bit changed.
Some features such as `~` and `^` are removed.
Spaces ` ` aren't treated as `AND` operator. For `AND` operator, you have to use comma `,`.

👎 

```
>= 1.2 < 3.0.0
```

👍 

```
>= 1.2, < 3.0.0
```

We think almost expressions are compatible with new specification.

## New specification of `semver`

```
<operator> <version>[, <operator> <version> ...]
```

Supported operators

- `>=`
- `<=`
- `!=`
- `>`
- `<`
- `=`

Support multiple constraints separated with comma `,`.
Multiple constraints are evaluated as `AND` comparison.
Spaces are trimmed.

## Why this change is needed

Prerelease versions aren't compared properly in `version_constraint` and `version_filter`.
The comparison with prerelease versions is already evaluated as `false`.
For example, we expect `v1.1.0-1` is greater than `v1.0.0`, but the evaluation result of `version_constraint` is `false`.

This is due to third party library aqua depends on.
aqua uses [hashicorp/go-version](https://github.com/hashicorp/go-version)'s [Constraints#Check](https://pkg.go.dev/github.com/hashicorp/go-version#Constraints.Check).
This function has the above bug.
This is obviously a bug and there was [a bug report](https://github.com/hashicorp/go-version/issues/101), but maintainers closed the issue and seems not to fix it.

So I considered migrating hashicorp/go-version to another library such as [Masterminds/semver](https://github.com/Masterminds/semver).
But unfortunately, [Masterminds/semver](https://github.com/Masterminds/semver) has a same issue and the maintainer tells it's not bug.

- https://github.com/Masterminds/semver#checking-version-constraints
- https://github.com/Masterminds/semver#working-with-prerelease-versions

So I'm considering reimplementing `semver` and `semverWithVersion` ourselves.
For aqua, some operators such as `~` and `^` aren't needed.
