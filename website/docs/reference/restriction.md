---
sidebar_position: 1200
---

# Restriction

aqua has some restrictions.

## aqua doesn't support running any external commands to install tools

aqua's install process is very simple.

1. Download tool
1. Unarchive tool in $AQUA_ROOT_DIR

aqua doesn't support running external commands to install tools, though there are some exceptions such as [cosign, slsa-verifier](/docs/reference/security/cosign-slsa), [minisign](/docs/reference/security/minisign), [gh](/docs/reference/security/github-artifact-attestations), [go](/docs/reference/registry-config/go-install-package), and [cargo](/docs/reference/registry-config/cargo-package).
So aqua can't support tools requiring to run external commands.

This is not necessarily a draw back.

https://github.com/aquaproj/aqua-registry/issues/987#issuecomment-1104422712

> You may think it's inconvenient, but we think this design is important to keep aqua simple, secure, less dependency, and maintainable.
> 
> aqua doesn't need any dependency.
> aqua doesn't run external commands.
> aqua doesn't change files outside install directory.
> 
> So the trouble shooting is relatively easy.
> Otherwise, user support would be very hard.
