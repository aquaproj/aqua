---
sidebar_position: 500
---

# rosetta2

[#442](https://github.com/aquaproj/aqua/pull/442) [#444](https://github.com/aquaproj/aqua/pull/444)

If a package isn't built for apple silicon (i.e. `GOOS=darwin, GOARCH=arm64`), you have to install the package built for amd64 (i.e. `GOOS=darwin, GOARCH=amd64`).
With the field `rosetta2`, you don't have to write `if` condition to support such a case.
`rosetta2` must be boolean. By default, `rosetta2` is `false`.

If `rosetta2` is `true` and `GOOS` is `darwin` and `GOARCH` is `arm64`, the template variable `Arch` is interpreted as `GOARCH=amd64`.

AS IS

```yaml
asset: 'argo-{{.OS}}-{{if eq .GOOS "darwin"}}amd64{{else}}{{.Arch}}{{end}}.gz'
```

TO BE

```yaml
rosetta2: true
asset: 'argo-{{.OS}}-{{.Arch}}.gz'
```
