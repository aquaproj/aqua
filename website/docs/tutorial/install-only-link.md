---
sidebar_position: 400
---

# `aqua i`'s `-l` option

You added [tfmigrator/cli](https://github.com/tfmigrator/cli) in [Search packages](search-packages.md), but it isn't installed yet.

```bash
command -v tfmigrator # command is not found
```

Let's run `aqua i -l`.

```bash
aqua i -l
```

The command would exit immediately because the tool isn't downloaded and installed yet.

The command `aqua i` installs all tools at once.
But when the option `-l` is set, `aqua i` creates only symbolic links in `${AQUA_ROOT_DIR}/bin` and skips downloading and installing tools.

Even if downloading and installing are skipped, you can execute the tool thanks for [Lazy Install](lazy-install.md).

```bash
tfmigrator -v
```

`-l` option is useful for local development because you can install only tools which are needed for you.
