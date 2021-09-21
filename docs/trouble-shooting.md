# Trouble Shooting

When aqua doesn't work, the environment variable `AQUA_LOG_LEVEL` is helpful for the debug.

For example,

```console
$ AQUA_LOG_LEVEL=debug kind --help
DEBU[0000] CLI args                                      config= log_level=debug program=aqua
DEBU[0000] install the package                           package_name=kubernetes-sigs/kind package_version=v0.11.1 program=aqua registry=standard
DEBU[0000] check if the package is already installed     package_name=kubernetes-sigs/kind package_version=v0.11.1 program=aqua registry=standard
DEBU[0000] check the permission                          file_name=kind
DEBU[0000] execute the command                           exe_path=/Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64 program=aqua
DEBU[0000] command was executed but it failed            error="fork/exec /Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64: exec format error" exe_path=/Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64 exit_code=-1 program=aqua
DEBU[0000] command failed                                error="fork/exec /Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64/kind-darwin-amd64: exec format error" exit_code=-1 program=aqua
```

In the above case, kind is already installed but the `exec format error` occured. I reinstalled kind then the problem has been solved.

```
# uninstall kind
$ rm -R /Users/shunsuke-suzuki/.aqua/pkgs/http/kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64
# kind is reinstalled by Lazy Install
$ kind --help
```
