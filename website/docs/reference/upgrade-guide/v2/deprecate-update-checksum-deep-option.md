---
sidebar_position: 500
---

# Deprecate `update-checksum`'s `--deep` option and change the default behaviour same as `--deep` option

[#1769](https://github.com/aquaproj/aqua/issues/1769) [#1770](https://github.com/aquaproj/aqua/pull/1770) 

## Feature Overview

Change the default behaviour of `update-checksum` same as `--deep` option.

We keep `--deep` option in aqua v2 for the compatibility.
We will remove `--deep` option in aqua v3.

## Why is the feature needed?

To make `update-checksum` simple.
If checksum file isn't provided `aqua update-checksum` doesn't get checksums.
This behaviour is confusing and bothering.

### How to migrate

Please stop using `--deep` option.
aqua v2 still keeps `--deep` option but this option doesn't change anything.
This option would be removed in aqua v3.
