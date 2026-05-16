---
sidebar_position: 900
---

# Experimental Feature

[#725](https://github.com/aquaproj/aqua/issues/725), `aqua >= v1.6.0`

aqua supports the mechanism for experimental features.
They are disabled by default, but you can enable them with environment variables or somehow.

Maybe experimental features would become enabled by default, or maybe they would be aborted.
aqua conforms semantic versioning, so when there are breaking changes we release major update.
But experimental features are exception of semantic versioning, so maybe we abort them in the minor or patch update.

## AQUA_EXPERIMENTAL_X_SYS_EXEC

[#710](https://github.com/aquaproj/aqua/issues/710) [#715](https://github.com/aquaproj/aqua/pull/715) [#726](https://github.com/aquaproj/aqua/pull/726), `aqua >= v1.6.0`

:::caution
Deprecated in aqua v2.5.0.
Please see [AQUA_X_SYS_EXEC](/docs/reference/execve-2).
:::
