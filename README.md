# aqua

[![Build Status](https://github.com/suzuki-shunsuke/aqua/workflows/test/badge.svg)](https://github.com/suzuki-shunsuke/aqua/actions)
[![GitHub last commit](https://img.shields.io/github/last-commit/suzuki-shunsuke/aqua.svg)](https://github.com/suzuki-shunsuke/aqua)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/aqua/main/LICENSE)

Command line tool to install tools and manage their versions.

* [Tutorial](tutorial/README.md)
* [Usage](docs/usage.md)
* [Configuration](docs/config.md)

## Install

Please download a binary from the [Release Page](https://github.com/suzuki-shunsuke/aqua/releases).

Or you can install aqua quickly with [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer).

e.g.

```
$ curl -sSfL https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/v0.1.0/aqua-installer | bash
```

GitHub Actions

e.g.

```yaml
- uses: suzuki-shunsuke/aqua-installer@v0.1.0
  with:
    version: v0.1.0-6
    install_path: /tmp/bin/aqua
```

## Directory Structure

```
.aqua/bin/
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

## Related Projects

* [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy)
* [aqua-installer](https://github.com/suzuki-shunsuke/aqua-installer)

## Example

* [suzuki-shunsuke/my-aqua-config](https://github.com/suzuki-shunsuke/my-aqua-config)
* [suzuki-shunsuke/example-aqua](https://github.com/suzuki-shunsuke/example-aqua)

## License

[MIT](LICENSE)
