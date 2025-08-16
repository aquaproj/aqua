---
sidebar_position: 1960
---

# complete_windows_ext

aqua >= v1.12.0

Many Windows executable files have the suffix `.exe`, so aqua completes the suffix automatically.

aqua completes the suffix `.exe` to the following attributes.

* asset
* url
* files.src

Using the attribute `complete_windows_ext`, you can specify if `.exe` is completed.

### Example

```yaml
    files:
      - name: gh
        src: bin/gh # bin/gh.exe
```

```yaml
    format: raw
    asset: aws-vault-{{.OS}}-{{.Arch}} # aws-vault-{{.OS}}-{{.Arch}}.exe
```

```yaml
    complete_windows_ext: false # disable completion
    format: raw
    asset: aws-vault-{{.OS}}-{{.Arch}} # aws-vault-{{.OS}}-{{.Arch}}
```

```yaml
    url: https://storage.googleapis.com/container-diff/{{.Version}}/container-diff-{{.OS}}-amd64 # .exe is completed
    files:
      - name: container-diff
        src: container-diff-{{.OS}}-amd64 # container-diff-{{.OS}}-amd64.exe
```
