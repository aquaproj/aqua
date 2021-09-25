# Where are tools installed?

* Symbolic links are created in `$HOME/.aqua/bin`, so add this to the environment variable `PATH`
* Tools are installed in `$HOME/.aqua/pkgs`

```
(your working directory)/
  aqua.yaml
~/.aqua/ # $AQUA_ROOT_DIR (default ~/.aqua)
  bin/
    aqua-proxy (symbolic link to aqua-proxy)
    <tool> (symbolic link to aqua-proxy)
  global/
    aqua.yaml # global configuration
  pkgs/
    github_release/
      github.com/
        suzuki-shunsuke/
          aqua-proxy/
            v0.1.0/
              aqua-proxy_darwin_amd64.tar.gz
                aqua-proxy
  registries/
    github_content/
      github.com/
        suzuki-shunsuke/
          aqua-registry/
            v0.1.1-0/
              registry.yaml
```
