---
sidebar_position: 810
---

# Progress Bar

aqua >= [v1.17.0](https://github.com/aquaproj/aqua/releases/tag/v1.17.0)

The progress bar was introduced at [v1.15.0](https://github.com/aquaproj/aqua/releases/tag/v1.15.0), and was disabled by default at [v1.17.0](https://github.com/aquaproj/aqua/releases/tag/v1.17.0).

* v1.15.0 [#956](https://github.com/aquaproj/aqua/issues/956) [#963](https://github.com/aquaproj/aqua/pull/963)
* v1.17.0 [#976](https://github.com/aquaproj/aqua/issues/976) [#979](https://github.com/aquaproj/aqua/pull/979)

The progress bar is disabled by default, because the progress bar isn't essential but it may cause some trouble depending on the environment.

You can enable it by setting the environment variable `AQUA_PROGRESS_BAR` to `true`.

```console
$ export AQUA_PROGRESS_BAR=true
```

The progress bar is powered by [github.com/schollz/progressbar/v3](https://pkg.go.dev/github.com/schollz/progressbar/v3).

e.g.

![image](https://user-images.githubusercontent.com/13323303/176585183-b8616482-5e3b-4f99-be98-6e7d752c5dbc.png)

```console
$ aqua i
INFO[0000] download and unarchive the package            aqua_version= env=darwin/arm64 package_name=aqua-proxy package_version=v1.1.2 program=aqua registry=
Downloading aqua-proxy v1.1.2 100% |██████████████████████████████████████████████████████████████████████████████████████████████████████████| (668/668 kB, 11.270 MB/s)
INFO[0000] create a symbolic link                        aqua_version= env=darwin/arm64 link_file=/Users/shunsukesuzuki/Documents/test/aqua/progress/aquaproj-aqua/bin/aqua-proxy new=../pkgs/github_release/github.com/aquaproj/aqua-proxy/v1.1.2/aqua-proxy_darwin_arm64.tar.gz/aqua-proxy package_name=aqua-proxy package_version=v1.1.2 program=aqua registry=
INFO[0000] create a symbolic link                        aqua_version= env=darwin/arm64 link_file=/Users/shunsukesuzuki/Documents/test/aqua/progress/aquaproj-aqua/bin/aqua-installer new=aqua-proxy package_name=aquaproj/aqua-installer package_version=v1.0.0 program=aqua registry=standard registry_ref=v3.4.0
INFO[0000] create a symbolic link                        aqua_version= env=darwin/arm64 link_file=/Users/shunsukesuzuki/Documents/test/aqua/progress/aquaproj-aqua/bin/tokei new=aqua-proxy package_name=XAMPPRocky/tokei package_version=v12.1.2 program=aqua registry=standard registry_ref=v3.4.0
INFO[0000] create a symbolic link                        aqua_version= env=darwin/arm64 link_file=/Users/shunsukesuzuki/Documents/test/aqua/progress/aquaproj-aqua/bin/terraform new=aqua-proxy package_name=hashicorp/terraform package_version=v1.2.3 program=aqua registry=standard registry_ref=v3.4.0
INFO[0000] create a symbolic link                        aqua_version= env=darwin/arm64 link_file=/Users/shunsukesuzuki/Documents/test/aqua/progress/aquaproj-aqua/bin/tfcmt new=aqua-proxy package_name=suzuki-shunsuke/tfcmt package_version=v3.2.5 program=aqua registry=standard registry_ref=v3.4.0
INFO[0000] download and unarchive the package            aqua_version= env=darwin/arm64 package_name=suzuki-shunsuke/tfcmt package_version=v3.2.5 program=aqua registry=standard
INFO[0000] download and unarchive the package            aqua_version= env=darwin/arm64 package_name=XAMPPRocky/tokei package_version=v12.1.2 program=aqua registry=standard
INFO[0000] download and unarchive the package            aqua_version= env=darwin/arm64 package_name=aquaproj/aqua-installer package_version=v1.0.0 program=aqua registry=standard
INFO[0000] download and unarchive the package            aqua_version= env=darwin/arm64 package_name=hashicorp/terraform package_version=v1.2.3 program=aqua registry=standard
Downloading aquaproj/aqua-installer v1.0.0   0% |                                                                                                            | ( 0/ 0B, ) [0s:0s]
Downloading XAMPPRocky/tokei v12.1.2 100% |████████████████████████████████████████████████████████████████████████████████████████████████████| (1.5/1.5 MB, 7.936 MB/s)
Downloading hashicorp/terraform v1.2.3 100% |███████████████████████████████████████████████████████████████████████████████████████████████████| (19/19 MB, 23.924 MB/s)
Downloading suzuki-shunsuke/tfcmt v3.2.5 100% |████████████████████████████████████████████████████████████████████████████████████████████████| (3.7/3.7 MB, 2.076 MB/s)
```
