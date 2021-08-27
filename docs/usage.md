# Usage

```console
$ aqua help
NAME:
   aqua - General version manager. https://github.com/suzuki-shunsuke/aqua

USAGE:
   aqua [global options] command [command options] [arguments...]

VERSION:
   0.1.0-7 (unreleased)

COMMANDS:
   install, i   Install tools
   exec         Execute tool
   get-bin-dir  Show the path where the packages are installed
   version      Show version
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-level value         log level [$AQUA_LOG_LEVEL]
   --config value, -c value  configuration file path [$AQUA_CONFIG]
   --help, -h                show help (default: false)
   --version, -v             print the version (default: false)
```

## aqua install

```console
$ aqua help install
NAME:
   aqua install - Install tools

USAGE:
   aqua install [command options] [arguments...]

OPTIONS:
   --only-link  create links but skip download packages (default: false)
   --help, -h   show help (default: false)
```

## aqua exec

```console
$ aqua help exec   
NAME:
   aqua exec - Execute tool

USAGE:
   aqua exec [arguments...]
```

## aqua get-bin-dir

```console
$ aqua help get-bin-dir
NAME:
   aqua get-bin-dir - Show the path where the packages are installed

USAGE:
   aqua get-bin-dir [arguments...]
```
