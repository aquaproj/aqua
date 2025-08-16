---
sidebar_position: 350
---

# aqua-proxy

https://github.com/aquaproj/aqua-proxy

The internal CLI tool of aqua.

We develop aqua-proxy for aqua, and we don't assume that aqua-proxy is used in the other purpose.

Basically the user of aqua don't have to know the detail of aqua-proxy.
aqua-proxy is installed to `$AQUA_ROOT_DIR/bin/aqua-proxy` automatically when `aqua install` and `aqua exec` is run, so you don't have to install aqua-proxy explicitly.

aqua-proxy has only the minimum feature and responsibility.
aqua-proxy is stable and isn't changed basically.

aqua-proxy is developed to decide the version of aqua and package managed with aqua dynamically according to the aqua's configuration file when the package is executed.

Please see [How does Lazy Install work?](/docs/reference/lazy-install) too.
