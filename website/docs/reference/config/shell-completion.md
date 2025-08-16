---
sidebar_position: 890
---

# Shell Completion

- `aqua >= v1.11.0` [#856](https://github.com/aquaproj/aqua/issues/856) [#859](https://github.com/aquaproj/aqua/pull/859)
- Fish shell: aqua >= [v2.29.0](https://github.com/aquaproj/aqua/releases/tag/v2.29.0) [#2929](https://github.com/aquaproj/aqua/pull/2929)

The commands `aqua completion <SHELL>` outputs shell completion code.
By sourcing the output of `aqua completion <SHELL>` in .bashrc, .zshrc, or ~/.config/fish/completions/aqua.fish,
aqua supports the shell completion.

.bashrc

```sh
if command -v aqua &> /dev/null; then
    eval "$(aqua completion bash)"
fi
```

.zshrc

```sh
if command -v aqua &> /dev/null; then
    source <(aqua completion zsh)
fi
```

https://fishshell.com/docs/current/completions.html#where-to-put-completions

```sh
aqua completion fish > ~/.config/fish/completions/aqua.fish
```

## Bash Completion

```console
$ aqua # Enter <Tab> Key
completion  exec        g           generate    h           help        i           init        install     list        version     which
```

```console
$ aqua i  # Enter <Tab> Key
i        init     install
```

## Zsh Completion

```console
$ aqua # Enter <Tab> Key
completion     -- Output shell completion script for bash or zsh
exec           -- Execute tool
generate    g  -- Search packages in registries and output the configuration interactively
help        h  -- Shows a list of commands or help for one command
init           -- Create a configuration file if it doesn't exist
install     i  -- Install tools
list           -- List packages in Registries
version        -- Show version
which          -- Output the absolute file path of the given command
```

```console
$ aqua i # Enter <Tab> Key
init        -- Create a configuration file if it doesn't exist
install  i  -- Install tools
```

```console
$ aqua i - # Enter <Tab> Key
--all        --help       --only-link  --test       -a           -h           -l
```
