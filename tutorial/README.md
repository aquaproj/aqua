# Tutorial

## Requirements

* Docker Engine
* Docker Compose
* Git

aqua doesn't depends on these tools, but in the tutorial you use them.

And the environment variable `GITHUB_TOKEN`, which is a [GitHub Access Token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token), is required. The required permission is `repo`.

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
```

Please see `aqua.yaml`.

```yaml
registries:
- type: standard
  ref: v0.8.0 # renovate: depName=suzuki-shunsuke/aqua-registry
packages:
- name: cli/cli
  registry: standard
  version: v2.0.0 # renovate: depName=cli/cli
```

In the tutorial, you will install [gh](https://cli.github.com/) and switch the version of gh with aqua.

Let's install tools with aqua.

```console
bash-5.1# aqua install
INFO[0000] download and unarchive the package            package_name=aqua-proxy package_version=v0.1.2 program=aqua registry=inline
INFO[0001] create a symbolic link                        link_file=/root/.aqua/bin/gh new=aqua-proxy program=aqua
INFO[0001] download and unarchive the package            package_name=cli/cli package_version=v2.0.0 program=aqua registry=standard
```

In addition to gh, [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy) is installed. aqua-proxy is required for aqua.

You can execute the installed tool by `aqua exec` command.

```console
bash-5.1# aqua exec -- gh version
gh version 2.0.0 (2021-08-24)
https://github.com/cli/cli/releases/tag/v2.0.0
```

By adding `$HOME/.aqua/bin` to the environment variable `PATH`, you can execute the installed tool directly.

```sh
export PATH=$HOME/.aqua/bin:$PATH
```

In this tutorial, the environment variable is already set in docker-compose.yml.

```console
bash-5.1# gh version
gh version 2.0.0 (2021-08-24)
https://github.com/cli/cli/releases/tag/v2.0.0
```

Please check `~/.aqua`.

```console
bash-5.1# tree -L 11 ~/.aqua
/root/.aqua
├── bin
│   ├── aqua-proxy -> ../pkgs/github_release/github.com/suzuki-shunsuke/aqua-proxy/v0.1.2/aqua-proxy_linux_amd64.tar.gz/aqua-proxy
│   └── gh -> aqua-proxy
├── pkgs
│   └── github_release
│       └── github.com
│           ├── cli
│           │   └── cli
│           │       └── v2.0.0
│           │           └── gh_2.0.0_linux_amd64.tar.gz
│           │               └── gh_2.0.0_linux_amd64
│           │                   ├── LICENSE
│           │                   ├── bin
│           │                   │   └── gh
│           │                   └── share
│           │                       └── man
│           │                           └── man1
│           └── suzuki-shunsuke
│               └── aqua-proxy
│                   └── v0.1.2
│                       └── aqua-proxy_linux_amd64.tar.gz
│                           ├── LICENSE
│                           ├── README.md
│                           └── aqua-proxy
└── registries
    └── github_content
        └── github.com
            └── suzuki-shunsuke
                └── aqua-registry
                    └── v0.1.1-0
                        └── registry.yaml

23 directories, 8 files
```

`$HOME/.aqua/bin/gh` is a symbolic link to [aqua-proxy](https://github.com/suzuki-shunsuke/aqua-proxy).

Run `aqua install` again, then the command exits immediately because tools are already installed properly.

```console
bash-5.1# aqua install
bash-5.1#
```

The subcommand `install` is a little long. You can use the short alias `i`.

```console
bash-5.1# aqua i
bash-5.1#
```

Please edit `aqua.yaml` to change gh's version to v1.14.0.

Run `aqua i` again, then gh v1.14.0 is installed.

```console
bash-5.1# aqua i
INFO[0000] download and unarchive the package            package_name=gh package_version=v1.14.0 program=aqua registry=standard
```

```console
bash-5.1# tree -L 11 ~/.aqua
/root/.aqua
├── bin
│   ├── aqua-proxy -> ../pkgs/github_release/github.com/suzuki-shunsuke/aqua-proxy/v0.1.2/aqua-proxy_linux_amd64.tar.gz/aqua-proxy
│   └── gh -> aqua-proxy
├── pkgs
│   └── github_release
│       └── github.com
│           ├── cli
│           │   └── cli
│           │       ├── v1.14.0
│           │       │   └── gh_1.14.0_linux_amd64.tar.gz
│           │       │       └── gh_1.14.0_linux_amd64
│           │       │           ├── LICENSE
│           │       │           ├── bin
│           │       │           │   └── gh
│           │       │           └── share
│           │       │               └── man
│           │       │                   └── man1
│           │       └── v2.0.0
│           │           └── gh_2.0.0_linux_amd64.tar.gz
│           │               └── gh_2.0.0_linux_amd64
│           │                   ├── LICENSE
│           │                   ├── bin
│           │                   │   └── gh
│           │                   └── share
│           │                       └── man
│           │                           └── man1
│           └── suzuki-shunsuke
│               └── aqua-proxy
│                   └── v0.1.2
│                       └── aqua-proxy_linux_amd64.tar.gz
│                           ├── LICENSE
│                           ├── README.md
│                           └── aqua-proxy
└── registries
    └── github_content
        └── github.com
            └── suzuki-shunsuke
                └── aqua-registry
                    └── v0.1.1-0
                        └── registry.yaml

30 directories, 10 files
```

```console
bash-5.1# gh version
gh version 1.14.0 (2021-08-04)
https://github.com/cli/cli/releases/tag/v1.14.0
```

Please edit `aqua.yaml` to change the version v2.0.0.
gh v2.0.0 is already installed, so the version is changed immediately.

```console
bash-5.1# gh version
gh version 2.0.0 (2021-08-24)
https://github.com/cli/cli/releases/tag/v2.0.0
```

Please edit `aqua.yaml` to change the version to v1.13.1 and run `gh version`.
gh v1.13.1 isn't installed yet, but it is installed automatically and is run.
You don't have to run `aqua i` in advance.

```console
bash-5.1# gh version
INFO[0000] download and unarchive the package            package_name=cli/cli package_version=v1.13.1 program=aqua registry=standard
gh version 1.13.1 (2021-07-20)
https://github.com/cli/cli/releases/tag/v1.13.1
```
