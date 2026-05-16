---
sidebar_position: 430
---

# Windows Support

`aqua >= v1.12.0`

[#850](https://github.com/aquaproj/aqua/issues/850)
[Project#4](https://github.com/orgs/aquaproj/projects/4)

The author [@suzuki-shunsuke](https://github.com/suzuki-shunsuke) doesn't use Windows, so the help of Windows Users is welcome.

## Install

Please see [Install](/docs/install).

## The separator of AQUA_GLOBAL_CONFIG

* Command Prompt, PowerShell: `;`
* Git Bash: `:`

## Required Standard Registry Version

Please upgrade [Standard Registry](https://github.com/aquaproj/aqua-registry) to v2.28.1 or later.

## Windows Settings

aqua works even if `Developer Mode` is disabled.
And you don't have to run the terminal as Administrator.

### Windows Security

:::caution
Please change the settings at your own risk.
:::

Security software may prevent aqua from installing and running tools.
In that case, you may have to add `AQUA_ROOT_DIR` to security software's exclusion. 

## Windows Support of installed tools

Note that some tools don't support Windows.
aqua skips installing those tools on Windows with [supported_if](/docs/reference/registry-config/supported-if) or [supported_envs](/docs/reference/registry-config/supported-envs).

### tools written in shell scripts aren't supported

Currently, tools written in shell scripts aren't supported.

## Windows specific features

### Auto completion of the file extension

Please see [complete_windows_ext](/docs/reference/registry-config/complete-windows-ext).

### Create hard links instead of symbolic links

[#2918](https://github.com/aquaproj/aqua/issues/2918) aqua >= v2.30.0

Please see [Reference](/docs/reference/lazy-install#on-windows).

### Windows ARM Emulation

Please see [windows_arm_emulation](/docs/reference/registry-config/windows_arm_emulation).

## Trouble Shooting

### Interactive Search by `aqua g` doesn't work on Git Bash

The guide of `gh auth login` is helpful.

https://github.com/ktr0731/go-fuzzyfinder/issues/2#issuecomment-1154872091

> You appear to be running in MinTTY without pseudo terminal support.
> 
> MinTTY is the terminal emulator that comes by default with Git
> for Windows. It has known issues with gh's ability to prompt a
> user for input.
> 
> There are a few workarounds to make gh work with MinTTY:
> 
> - Reinstall Git for Windows, checking "Enable experimental support for pseudo consoles".
> 
> ![image](https://user-images.githubusercontent.com/13323303/173531978-21a99818-11ff-4385-962a-64f74e4023db.png)
> 
> - Use a different terminal emulator with Git for Windows like Windows Terminal.
>   You can run "C:\Program Files\Git\bin\bash.exe" from any terminal emulator to continue
>   using all of the tooling in Git For Windows without MinTTY.
> 
> - Prefix invocations of gh with winpty, eg: "winpty gh auth login".
>   NOTE: this can lead to some UI bugs.
