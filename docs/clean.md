# How to clean unused packages

aqua installs packages in `~/.aqua/pkgs` and doesn't remove them.
Even if you change the package version, aqua doesn't remove the old package.
If `~/.aqua/pkgs` consumes the disk usage and you want it to be slim, you can remove packages in `~/.aqua/pkgs` by yourself.

The simplest way is to remove `~/.aqua`.

```
$ rm -R ~/.aqua
```

By keeping `~/.aqua/bin`, you can install tools by the Lazy Install without running `aqua i`.

You can also remove the specific package or package version.

```
$ rm -R ~/.aqua/pkgs/github_release/github.com/suzuki-shunsuke/akoi
```

```
$ rm -R ~/.aqua/pkgs/github_release/github.com/suzuki-shunsuke/akoi/v2.2.0
```

```
$ rm -R ~/.aqua/pkgs/github_release/github.com/suzuki-shunsuke
```
