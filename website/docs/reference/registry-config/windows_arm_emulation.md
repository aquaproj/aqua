---
sidebar_position: 510
---

# windows_arm_emulation

[#2514](https://github.com/orgs/aquaproj/discussions/2514) [#2515](https://github.com/aquaproj/aqua/pull/2515)

aqua >= v2.20.0

ARM based Windows 11 supports the emulation to run x64 Windows apps.

https://learn.microsoft.com/en-us/windows/arm/add-arm-support#emulation-on-arm-based-devices-for-x86-or-x64-windows-apps

> Windows 11 extends that emulation to run unmodified x64 Windows apps on Arm-powered devices.

If the field `windows_arm_emulation` is true, aqua uses pre built binaries for Windows amd64 on Windows arm64.
`windows_arm_emulation` must be boolean. By default, `windows_arm_emulation` is `false`.

`windows_arm_emulation` is similar with [rosetta2](rosetta2.md).

If `windows_arm_emulation` is `true` and `GOOS` is `windows` and `GOARCH` is `arm64`, the template variable `Arch` is interpreted as `GOARCH=amd64`.
