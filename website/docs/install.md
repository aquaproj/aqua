---
sidebar_position: 150
mdx:
  format: mdx
---

import TabItem from '@theme/TabItem';
import InstallationSetup, {InstallationMethodTabs, InstallationShellTabs, InstallationTabLabel, InstallationShellLink} from '@site/src/components/InstallationSetup';

# Install

aqua is a single binary written in Go.

For your local terminal, complete all three steps:

1. **Install aqua** using one of the methods below.
2. **Save the `PATH` setting** in your shell configuration: <InstallationShellLink platform="linux" hash="#bash">Bash</InstallationShellLink>, <InstallationShellLink platform="macos" hash="#zsh">Zsh</InstallationShellLink>, or <InstallationShellLink platform="windows" hash="#powershell">Windows</InstallationShellLink>. Running `export` only in your terminal does not persist the setting.
3. **Open a new terminal** and run `aqua -v` to check that aqua is still available.

You can also enable [shell completion](/docs/reference/config/shell-completion) after setup.

## 1. Install the binary `aqua` in `PATH`

Choose one installation method. For a local terminal, continue to step 2 to save your `PATH` setting.

<InstallationMethodTabs>
<TabItem value="script" label={<InstallationTabLabel title="Shell Script" detail="Linux / macOS / WSL" />}>

For Linux, macOS, and WSL.

### aqua-installer (Shell Script)

Run the installer, then continue to step 2 below to save your `PATH` setting.

```bash
curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer
echo "98b883756cdd0a6807a8c7623404bfc3bc169275ad9064dc23a6e24ad398f43d  aqua-installer" | sha256sum -c -
chmod +x aqua-installer
./aqua-installer
```

On macOS, use `shasum -a 256 -c -` instead of `sha256sum -c -`.
See [aqua-installer](/docs/products/aqua-installer#shell-script) for installer options.

:::tip Next: save your PATH setting
Next, [save the `PATH` setting](#2-set-the-environment-variable-path) so your shell can find tools installed by aqua.
You still need to complete step 2 to save this PATH setting for new terminal sessions.
:::

</TabItem>

<TabItem value="homebrew" label={<InstallationTabLabel title="Homebrew" detail="macOS / Linux" />}>

For macOS and Linux with Homebrew installed.

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

:::tip Next: save your PATH setting
Next, [save the `PATH` setting](#2-set-the-environment-variable-path) so your shell can find tools installed by aqua.
You still need to complete step 2 to save this PATH setting for new terminal sessions.
:::

</TabItem>

<TabItem value="winget" label={<InstallationTabLabel title="Winget" detail="Windows" />}>

For Windows. Run the command in PowerShell or Command Prompt.

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

:::tip Next: save your PATH setting
Next, [save the `PATH` setting](#2-set-the-environment-variable-path) so your shell can find tools installed by aqua.
You still need to complete step 2 to save this PATH setting for new terminal sessions.
:::

</TabItem>

<TabItem value="scoop" label={<InstallationTabLabel title="Scoop" detail="Windows" />}>

For Windows with Scoop installed.

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

:::tip Next: save your PATH setting
Next, [save the `PATH` setting](#2-set-the-environment-variable-path) so your shell can find tools installed by aqua.
You still need to complete step 2 to save this PATH setting for new terminal sessions.
:::

</TabItem>

<TabItem value="go" label={<InstallationTabLabel title="Go" detail="Go toolchain" />}>

Requires Go. Available on Linux, macOS, Windows, and WSL.

### go install

```sh
go install github.com/aquaproj/aqua/v2/cmd/aqua@latest
```

:::tip Next: save your PATH setting
Next, [save the `PATH` setting](#2-set-the-environment-variable-path) so your shell can find tools installed by aqua.
You still need to complete step 2 to save this PATH setting for new terminal sessions.
:::

</TabItem>

<TabItem value="binary" label={<InstallationTabLabel title="Prebuilt binaries" detail="GitHub Releases" />}>

### Download prebuilt binaries from GitHub Releases

https://github.com/aquaproj/aqua/releases

<details>
<summary>Verify downloaded binaries from GitHub Releases</summary>

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

</details>

:::tip Next: save your PATH setting
Next, [save the `PATH` setting](#2-set-the-environment-variable-path) so your shell can find tools installed by aqua.
You still need to complete step 2 to save this PATH setting for new terminal sessions.
:::

</TabItem>
<TabItem value="github-actions" label={<InstallationTabLabel title="GitHub Actions" detail="CI workflows" />}>

Add the following step to your workflow:

```yaml
- uses: aquaproj/aqua-installer@96a9bc20066c5bf5e275b41019cfc165b25f4e2e # v4.0.5
  with:
    aqua_version: v2.43.1
```

See [aqua-installer (GitHub Actions)](/docs/products/aqua-installer#github-actions) for configuration options.
:::info PATH is configured automatically
The action configures `PATH` for the workflow, so the local shell setup in steps 2 and 3 is not needed.
:::

</TabItem>
<TabItem value="circleci" label={<InstallationTabLabel title="CircleCI Orb" detail="CI workflows" />}>

Follow the [CircleCI Orb setup instructions](/docs/products/circleci-orb-aqua).

:::info Follow the integration setup
Use the linked instructions to configure CircleCI. Steps 2 and 3 below are for local terminal installations.
:::

</TabItem>
<TabItem value="devcontainer" label={<InstallationTabLabel title="Dev Container Feature" detail="Development containers" />}>

Follow the [Dev Container Feature setup instructions](https://github.com/aquaproj/devcontainer-features/tree/main/src/aqua-installer).

:::info Follow the integration setup
Use the linked instructions to configure your development container. Steps 2 and 3 below are for local terminal installations.
:::

</TabItem>
</InstallationMethodTabs>

## 2. Set the environment variable `PATH`

<InstallationSetup>

<InstallationShellTabs>
<TabItem value="linux" label={<InstallationTabLabel title="Bash" detail="Linux / WSL / Git Bash" />}>

#### Bash

##### Linux / WSL

```bash
cat >> ~/.bashrc <<'EOF'
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
EOF

source ~/.bashrc
```

Add the `export PATH=...` line shown above to `~/.bashrc` so aqua stays on PATH in new shells.
The commands above append this line and reload the file; run them once.


<details>
<summary>Git Bash on Windows</summary>

##### Git Bash (Windows)

```bash
cat >> ~/.bashrc <<'EOF'
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-${LOCALAPPDATA:-$HOME/AppData/Local}}/aquaproj-aqua}/bin:$PATH"
EOF

source ~/.bashrc
```

This appends the Windows-specific PATH setting to `~/.bashrc` and loads it into the current shell.
Git Bash uses the Windows install directory; WSL uses the Linux instructions above.
Run this once, or just run `source ~/.bashrc` if you have already saved the line.

The install directory is selected in this order: `AQUA_ROOT_DIR`, `XDG_DATA_HOME/aquaproj-aqua`, `LOCALAPPDATA/aquaproj-aqua`, then `$HOME/AppData/Local/aquaproj-aqua`.

</details>

Bash login shells read `~/.bash_profile`, `~/.bash_login`, or `~/.profile` instead of `~/.bashrc` (the first file that exists).
If your terminal starts a login shell, ensure that file sources `~/.bashrc`, or add the `export` line to that file too.

<details>
<summary>PATH options and other version managers</summary>

If you customize `AQUA_ROOT_DIR` or `XDG_DATA_HOME`, set it before the PATH line in your shell configuration.
The quoted `EOF` keeps the variables unchanged until your shell reads the configuration.

If aqua is already available in your shell, aqua v2.8.0 or later can also print its root directory:

```bash
export PATH="$(aqua root-dir)/bin:$PATH"
```

Save this line in your shell configuration if you use it instead of the example above.

If you use another version manager such as asdf, place aqua's PATH setting after that manager's setup.
See [using aqua with other tools](/docs/reference/use-aqua-with-other-tools).

</details>

</TabItem>
<TabItem value="macos" label={<InstallationTabLabel title="Zsh" detail="macOS" />}>

#### Zsh

```zsh
cat >> "${ZDOTDIR:-$HOME}/.zshrc" <<'EOF'
export PATH="${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}/bin:$PATH"
EOF

source "${ZDOTDIR:-$HOME}/.zshrc"
```

Add the `export PATH=...` line shown above to `${ZDOTDIR:-$HOME}/.zshrc` so aqua stays on PATH in new shells.
The commands above append this line and reload the file; run them once.


<details>
<summary>PATH options and other version managers</summary>

If you customize `AQUA_ROOT_DIR` or `XDG_DATA_HOME`, set it before the PATH line in your shell configuration.
The quoted `EOF` keeps the variables unchanged until your shell reads the configuration.

If aqua is already available in your shell, aqua v2.8.0 or later can also print its root directory:

```bash
export PATH="$(aqua root-dir)/bin:$PATH"
```

Save this line in your shell configuration if you use it instead of the example above.

If you use another version manager such as asdf, place aqua's PATH setting after that manager's setup.
See [using aqua with other tools](/docs/reference/use-aqua-with-other-tools).

</details>

</TabItem>
<TabItem value="windows" label={<InstallationTabLabel title="PowerShell" detail="Windows" />}>

#### PowerShell

```powershell
if (!(Test-Path -Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force
}
Add-Content -Path $PROFILE -Value 'Set-Item Env:Path "$Env:LOCALAPPDATA\aquaproj-aqua\bin;$Env:Path"'
. $PROFILE
```

PowerShell reads its profile when it starts. The commands above create that file if needed, append aqua's PATH setting, and load it into the current shell.
`Add-Content` preserves the existing contents. Run this once; if the line is already saved, only run `. $PROFILE` to reload it.
See [PowerShell profiles](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_profiles) for details about the startup file.

<details>
<summary>Custom install directory and PATH options</summary>

If `LOCALAPPDATA` is not set, use this line in your profile instead:

```powershell
Set-Item Env:Path "$Env:USERPROFILE\AppData\Local\aquaproj-aqua\bin;$Env:Path"
```

If you set `AQUA_ROOT_DIR`, use:

```powershell
Set-Item Env:Path "$Env:AQUA_ROOT_DIR\bin;$Env:Path"
```

Set any custom root directory before the PATH line in your profile.
If you use another version manager, place aqua's PATH setting after that manager's setup.
See [using aqua with other tools](/docs/reference/use-aqua-with-other-tools).

</details>

See [Windows support](/docs/reference/windows-support) for more details.

</TabItem>
<TabItem value="command-prompt" label={<InstallationTabLabel title="Command Prompt" detail="Windows" />}>

#### Command Prompt

1. Open **Edit environment variables for your account** from the Windows Start menu.
2. Under **User variables**, select **Path**, choose **Edit**, and add `%LOCALAPPDATA%\aquaproj-aqua\bin` as a new entry.
3. Save the change and open a new terminal.

This saves aqua's bin directory in your user PATH so new Command Prompt sessions can find aqua and its tools.
If you installed aqua under `AQUA_ROOT_DIR`, add that directory's `bin` folder instead.

<details>
<summary>Set PATH for the current session only</summary>

```batch
SET "PATH=%LOCALAPPDATA%\aquaproj-aqua\bin;%PATH%"
```

If `LOCALAPPDATA` is not set:

```batch
SET "PATH=%USERPROFILE%\AppData\Local\aquaproj-aqua\bin;%PATH%"
```

If you set `AQUA_ROOT_DIR`:

```batch
SET "PATH=%AQUA_ROOT_DIR%\bin;%PATH%"
```

These commands only update the current session. Use the steps above to persist the setting.

</details>

</TabItem>
</InstallationShellTabs>

</InstallationSetup>

## 3. Check your setup

<InstallationSetup>

Open a new terminal and confirm aqua is still available:

```sh
aqua -v
```

If the command is not found, check that you saved the `PATH` setting in the configuration file your shell reads.

</InstallationSetup>
