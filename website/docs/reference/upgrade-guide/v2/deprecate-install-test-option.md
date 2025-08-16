---
sidebar_position: 800
---

# install: Deprecate `--test` option and change the default behaviour same as one with `--test` option

[#1691](https://github.com/aquaproj/aqua/issues/1691) [#1694](https://github.com/aquaproj/aqua/pull/1694) 

Deprecate `--test` option and change the default behaviour same as one with `--test` option.
To keep the compatibility, `--test` option isn't removed in aqua v2 but the option doesn't change anything.
`--test` option will be removed in aqua v3.

## Why is the feature needed?

- Make the code simple
- We think the default behaviour should be same as one with `--test` option. `aqua i` should fail if an executable file isn't found
