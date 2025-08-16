---
sidebar_position: 150
---

# Install

aqua is a single binary written in Go.

1. Install the binary `aqua` in `PATH`
1. Set the environment variable `PATH`
1. [(Optional) Shell completion](/docs/reference/config/shell-completion)

## 1. Install the binary `aqua` in `PATH`

- [Homebrew](#homebrew)
- Windows
  - [Winget](#winget)
  - [Scoop](#scoop)
- [aqua-installer (Shell Script)](/docs/products/aqua-installer#shell-script)
- [aqua-installer (GitHub Actions)](/docs/products/aqua-installer#github-actions)
- [CircleCI Orb](/docs/products/circleci-orb-aqua)
- [go install](#go-install)
- [Dev Container Feature](https://github.com/aquaproj/devcontainer-features/tree/main/src/aqua-installer)
- [Download prebuilt binaries from GitHub Releases](#download-prebuilt-binaries-from-github-releases)

### Homebrew

You can install aqua using [Homebrew](https://brew.sh/).

[Homebrew Core Formula: aqua](https://formulae.brew.sh/formula/aqua)

```sh
brew install aqua
```

Or

```sh
brew install aquaproj/aqua/aqua
```

### Winget

From [aqua v2.17.4](https://github.com/aquaproj/aqua/releases/tag/v2.17.4), you can install aqua by [Winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/).

```sh
winget install aquaproj.aqua
```

:::caution
Due to the mechanism of Winget, it takes a few days at most until we can install the latest version after the latest version has been released.
Everytime a new version is released, we need to send a pull request to [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) and wait until the pull request is merged.
[The list of pull requests](https://github.com/microsoft/winget-pkgs/pulls?q=is%3Aopen+is%3Apr+author%3Asuzuki-shunsuke+aquaproj.aqua+in%3Atitle)
:::

### Scoop

From [aqua v2.16.2](https://github.com/aquaproj/aqua/releases/tag/v2.16.2), you can also install aqua by [Scoop](https://scoop.sh/).

[Main bucket](https://github.com/ScoopInstaller/Main):

```sh
scoop bucket add main
scoop install main/aqua
```

[Our bucket](https://github.com/aquaproj/scoop-bucket):

```sh
scoop bucket add aquaproj https://github.com/aquaproj/scoop-bucket
scoop install aqua
```

### go install

```sh
go install github.com/aquaproj/aqua/v2/cmd/aqua@latest
```

### Download prebuilt binaries from GitHub Releases

https://github.com/aquaproj/aqua/releases

#### Verify downloaded binaries from GitHub Releases

You can verify downloaded binaries using some tools.

1. [Cosign](https://github.com/sigstore/cosign)
1. [slsa-verifier](https://github.com/slsa-framework/slsa-verifier)
1. [GitHub CLI](https://cli.github.com/)

--

1. Cosign:

You can install Cosign by aqua.

```sh
aqua g -i sigstore/cosign
```

```sh
# Download assets from GitHub Releases.
gh release download -R aquaproj/aqua v2.34.0
# Verify a checksum file.
cosign verify-blob \
  --signature aqua_2.34.0_checksums.txt.sig \
  --certificate aqua_2.34.0_checksums.txt.pem \
  --certificate-identity-regexp 'https://github\.com/suzuki-shunsuke/go-release-workflow/\.github/workflows/release\.yaml@.*' \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  aqua_2.34.0_checksums.txt
```

Output:

```
Verified OK
```

After verifying the checksum, verify the artifact.

```sh
cat aqua_2.34.0_checksums.txt | sha256sum -c --ignore-missing -
```

2. slsa-verifier

You can install slsa-verifier by aqua.

```sh
aqua g -i slsa-framework/slsa-verifier
```

```sh
# Download assets from GitHub Releases.
gh release download -R aquaproj/aqua v2.34.0
# Verify an asset.
slsa-verifier verify-artifact aqua_darwin_arm64.tar.gz \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/aquaproj/aqua \
  --source-tag v2.34.0
```

Output:

```
Verified signature against tlog entry index 133024968 at URL: https://rekor.sigstore.dev/api/v1/log/entries/108e9186e8c5677af3bf58014b72ab1571f566855d27109b70403a96394003283d540765fc0e2c20
Verified build using builder "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.0.0" at commit 2f9cc345c3c49b9a0c8fcd9d8e1c461bbd8fd533
Verifying artifact aqua_darwin_arm64.tar.gz: PASSED

PASSED: SLSA verification passed
```

3. GitHub CLI

You can install GitHub CLI by aqua.

```sh
aqua g -i cli/cli
```

```sh
# Download assets from GitHub Releases.
gh release download -R aquaproj/aqua v2.35.0-1 -p aqua_darwin_arm64.tar.gz
# Verify an asset.
gh attestation verify aqua_darwin_arm64.tar.gz \
  -R aquaproj/aqua \
  --signer-workflow suzuki-shunsuke/go-release-workflow/.github/workflows/release.yaml
```

Output:

```
Loaded digest sha256:763c8d5e6b8585ebb9d9bab0ee1fcafd4a29c3e7f44a85ac77780bac3ca6fff1 for file://aqua_darwin_arm64.tar.gz
Loaded 1 attestation from GitHub API
✓ Verification succeeded!

sha256:763c8d5e6b8585ebb9d9bab0ee1fcafd4a29c3e7f44a85ac77780bac3ca6fff1 was attested by:
REPO                                 PREDICATE_TYPE                  WORKFLOW                                                               
suzuki-shunsuke/go-release-workflow  https://slsa.dev/provenance/v1  .github/workflows/release.yaml@7f97a226912ee2978126019b1e95311d7d15c97a
```

## 2. Set the environment variable `PATH`

:::info
From aqua v2.8.0, `aqua root-dir` command is available.

```bash
export PATH="$(aqua root-dir)/bin:$PATH"
```
:::

:::tip
If you use aqua combined with other version manager such as asdf,
please add `${AQUA_ROOT_DIR}/bin` to the environment variable `PATH` after other version manager.
For detail, please see [here](/docs/reference/use-aqua-with-other-tools).
:::

### Linux, macOS

```sh
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
```

### Windows

About Windows, please see [here](/docs/reference/windows-support) too.

- Git Bash (mingw)
- PowerShell
- Command Prompt

#### Git Bash (mingw)

```sh
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-${LOCALAPPDATA:-$HOME/AppData/Local}}/aquaproj-aqua}/bin:$PATH"
```

The order of priority is as follows:

1. `$AQUA_ROOT_DIR/bin`: If `$AQUA_ROOT_DIR` is set
1. `$XDG_DATA_HOME/aquaproj-aqua/bin`: If `$XDG_DATA_HOME` is set
1. `$LOCALAPPDATA/aquaproj-aqua/bin`: If `$LOCALAPPDATA` is set
1. `$HOME/AppData/Local/aquaproj-aqua/bin`

#### PowerShell

```sh
Set-Item Env:Path "$Env:LOCALAPPDATA\aquaproj-aqua\bin;$Env:Path"
```

If `LOCALAPPDATA` isn't set,

```sh
Set-Item Env:Path "$Env:HOMEPATH\AppData\Local\aquaproj-aqua\bin;$Env:Path"
```

You can customize the path with the environment variable `AQUA_ROOT_DIR`.

```sh
Set-Item Env:Path "$Env:AQUA_ROOT_DIR\bin;$ENV:Path"
```

#### Command Prompt

```sh
SET PATH=%LOCALAPPDATA%\aquaproj-aqua\bin;%PATH%
```

If `LOCALAPPDATA` isn't set,

```sh
SET PATH=%HOMEPATH%\AppData\Local\aquaproj-aqua\bin;%PATH%
```

You can also customize the path with the environment variable `AQUA_ROOT_DIR`.

```sh
SET PATH=%AQUA_ROOT_DIR%\bin;%PATH%
```
