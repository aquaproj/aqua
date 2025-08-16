---
sidebar_position: 700
---

# `files`

- `name`: the file name
- `src`: (default: the value of `name`, type: `template string`) the path to the file from the archive file's root.
- `dir`: This is used only for go type package
- `link`: This is rarely used. aqua >= [v2.36.0](https://github.com/aquaproj/aqua/releases/tag/v2.36.0)
- `hard` (`boolean`, default: `false`) aqua >= [v2.49.0](https://github.com/aquaproj/aqua/releases/tag/v2.49.0)

## `link`

aqua >= [v2.36.0](https://github.com/aquaproj/aqua/releases/tag/v2.36.0)

`link` is used to change `$0` by symlink (hardlink on Windows).
Some tools change their behavior by `$0`.
For instance, [granted](https://github.com/common-fate/granted) changes the behavior based on `args[0]`.

https://github.com/common-fate/granted/blob/e8de3ec7d62d543062d8be802b27abb3d8fac429/cmd/granted/main.go#L37-L44

```go
	// Use a single binary to keep keychain ACLs simple, swapping behavior via argv[0]
	var app *cli.App
	switch filepath.Base(os.Args[0]) {
	case "assumego", "assumego.exe", "dassumego", "dassumego.exe":
		app = assume.GetCliApp()
	default:
		app = granted.GetCliApp()
	}
```

`link` allows you to change `$0` by symlink.

```yaml
files:
  - name: granted
  - name: assumego
    src: granted
    link: assumego # link is the relative path from src to the symlink
```

## `hard`

aqua >= [v2.49.0](https://github.com/aquaproj/aqua/releases/tag/v2.49.0) [#3775](https://github.com/aquaproj/aqua/issues/3775)

`link` is required.
If `hard` is true, aqua creates a hard link instead of a symbolic link.

```yaml
files:
  - name: pnpm
	link: pnpm
	hard: true
```
