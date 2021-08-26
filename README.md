# aqua

Command line tool to install tools and manage their versions.

## Install

Please download a binary from the [Release Page](https://github.com/suzuki-shunsuke/aqua/releases).

## Tutorial

[tutorial](tutorial/README.md)

## Directory Structure

```
.aqua/bin/ # config.bin_dir (default .aqua/bin)
  akoi (symbolic link to ~/.aqua/bin/aqua-proxy)
~/.aqua/ # $AQUA_ROOT_DIR (default ~/.aqua)
  bin/
    aqua-proxy (symbolic link to aqua-proxy)
  pkgs/
    github_release/
      github.com/
        suzuki-shunsuke/
          aqua-proxy/
            v0.1.0-0/
              aqua-proxy_darwin_amd64.tar.gz
                aqua-proxy
```

## License

[MIT](LICENSE)
