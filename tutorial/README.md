# Tutorial

## Requirements

* Docker Engine
* Docker Compose
* Git

aqua doesn't depends on these tools, but in the tutorial you use them.

And the environment variable `GITHUB_TOKEN`, which is a GitHub Access Token, is required.

## Tutorial

Clone this repository.

```
$ git clone https://github.com/suzuki-shunsuke/aqua
$ cd aqua/tutorial
```

Launch the Docker container.

```
$ docker-compose up -d
```

```
$ docker-compose exec main bash
```

aqua is installed in Dockerfile.

```console
bash-5.1# aqua -v
aqua version 0.1.0-3 (91ccad4ded9412504c305e07661aa6a43e2b5a91)
```

Please see `aqua.yaml`.

```yaml
packages:
- name: akoi
  repository: inline
  version: v2.2.0
inline_repository:
- name: akoi
  type: github_release
  repo: suzuki-shunsuke/akoi
  artifact: 'akoi_{{trimPrefix "v" .Package.Version}}_{{.OS}}_{{.Arch}}.tar.gz'
  files:
  - name: akoi
    src: akoi
```

In the tutorial, you will install [akoi](https://github.com/suzuki-shunsuke/akoi) and switch the version of akoi with aqua.

Let's install tools with aqua.

```console
bash-5.1# aqua install
INFO[0000] download and unarchive the package            package_name=aqua-proxy package_version=v0.1.0-0 repository=inline
INFO[0001] create a symbolic link                        link_file=/workspace/.aqua/bin/akoi new=/root/.aqua/bin/aqua-proxy
INFO[0001] download and unarchive the package            package_name=akoi package_version=v2.2.0 repository=inline
```

In addition to akoi, [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy) is installed. aqua-proxy is required for aqua.

You can execute the installed tool by `aqua exec` command.

```console
bash-5.1# aqua exec -- akoi help
NAME:
   akoi - binary version control system

USAGE:
   akoi [global options] command [command options] [arguments...]

VERSION:
   2.2.0

AUTHOR:
   suzuki-shunsuke https://github.com/suzuki-shunsuke

COMMANDS:
     init     create a configuration file if it doesn't exist
     install  intall binaries
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

By adding `.aqua/bin` to the environment variable `PATH`, you can execute the installed tool directly.

```sh
export PATH=$PWD/.aqua/bin:$PATH
```

In this tutorial, the environment variable is already set in docker-compose.yml.

```console
bash-5.1# akoi -v
akoi version 2.2.0
```

Please check `~/.aqua` and `.aqua`.

```console
bash-5.1# tree ~/.aqua
/root/.aqua
├── bin
│   └── aqua-proxy -> /root/.aqua/pkgs/github_release/github.com/suzuki-shunsuke/aqua-proxy/v0.1.0-0/aqua-proxy_linux_amd64.tar.gz/aqua-proxy
└── pkgs
    └── github_release
        └── github.com
            └── suzuki-shunsuke
                ├── akoi
                │   └── v2.2.0
                │       └── akoi_2.2.0_linux_amd64.tar.gz
                │           ├── LICENSE
                │           ├── README.md
                │           └── akoi
                └── aqua-proxy
                    └── v0.1.0-0
                        └── aqua-proxy_linux_amd64.tar.gz
                            ├── LICENSE
                            ├── README.md
                            └── aqua-proxy

11 directories, 7 files
```

```console
bash-5.1# tree .aqua
.aqua
└── bin
    └── akoi -> /root/.aqua/bin/aqua-proxy

1 directory, 1 file
```

`.aqua/bin/akoi` is a symbolic link to [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy).

Run `aqua install` again, then the command exits soon.

```console
bash-5.1# aqua install
bash-5.1#
```

The subcommand `install` is a little long. You can use the short alias `i`.

```console
bash-5.1# aqua i
bash-5.1#
```

Please edit `aqua.yaml` to change akoi's version from v2.2.0 to v2.2.1.

Run `aqua i` again, then akoi v2.2.1 is installed.

```console
bash-5.1# aqua i
INFO[0000] download and unarchive the package            package_name=akoi package_version=v2.2.1 repository=inline
```

```console
bash-5.1# tree ~/.aqua
/root/.aqua
├── bin
│   └── aqua-proxy -> /root/.aqua/pkgs/github_release/github.com/suzuki-shunsuke/aqua-proxy/v0.1.0-0/aqua-proxy_linux_amd64.tar.gz/aqua-proxy
└── pkgs
    └── github_release
        └── github.com
            └── suzuki-shunsuke
                ├── akoi
                │   ├── v2.2.0
                │   │   └── akoi_2.2.0_linux_amd64.tar.gz
                │   │       ├── LICENSE
                │   │       ├── README.md
                │   │       └── akoi
                │   └── v2.2.1
                │       └── akoi_2.2.1_linux_amd64.tar.gz
                │           ├── LICENSE
                │           ├── README.md
                │           └── akoi
                └── aqua-proxy
                    └── v0.1.0-0
                        └── aqua-proxy_linux_amd64.tar.gz
                            ├── LICENSE
                            ├── README.md
                            └── aqua-proxy

13 directories, 10 files
```

```console
bash-5.1# akoi -v
akoi version 2.2.1
```

Please edit `aqua.yaml` to change the version from v2.2.1 to v2.2.0.

And run `akoi -v`, then the version is changed soon.

```console
bash-5.1# akoi -v
akoi version 2.2.0
```

Please edit `aqua.yaml` to change the version to v2.1.0 and run `akoi`, then before akoi is run akoi v2.1.0 is installed automatically.
You don't have to run `aqua i` in advance.

```console
bash-5.1# akoi -v
INFO[0000] download and unarchive the package            package_name=akoi package_version=v2.1.0 repository=inline
akoi version 2.1.0
```
