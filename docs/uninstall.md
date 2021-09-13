# How to uninstall aqua

Remove `aqua` and `~/.aqua`.

e.g.

```console
$ rm /usr/local/bin/aqua
$ rm -R ~/.aqua
```

Unset the environment variables you set for aqua.

e.g.

```sh
export PATH=$HOME/.aqua/bin:$PATH # Remove `$HOME/.aqua/bin`
export AQUA_LOG_LEVEL=debug # Remove
```
