---
sidebar_position: 800
---

# Wrap aqua with task runner such as GNU Make

aqua is [easy to use](/docs/#easy-to-use), but generally speaking it is not easy to introduce a new tool and let developers install it in a large team and organization.

By wrapping aqua with a task runner such as [GNU Make](https://www.gnu.org/software/make/) and [Task](https://taskfile.dev/), you may be able to solve the issue.
If a task runner is already used in your project, it's easy to introduce aqua.
By hiding the setup of aqua (installing aqua, adding `PATH`, and running `aqua i [-l]`) from developers using task runner,
developers don't have to aware of aqua.

## Example

:::caution
We aren't familiar with GNU Make and Task. So the example code of Makefile and Taskfile.yml may not be good.
And this example doesn't work in Windows.
Your contribution is welcome.
:::

https://github.com/suzuki-shunsuke/poc-aqua-make

In this example, Terraform is managed by aqua and developers can run `terraform` via GNU Make or Task without awareness of aqua.

```console
$ make tf-init
```

<details>

```console
$ make tf-init
bash scripts/setup_aqua.sh
aqua-installer: OK
===> Installing aqua v2.2.3 for bootstrapping...
===> Downloading https://github.com/aquaproj/aqua/releases/download/v2.2.3/aqua_linux_arm64.tar.gz ...
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
100 5817k  100 5817k    0     0  5765k      0  0:00:01  0:00:01 --:--:-- 26.6M
===> Verifying checksum of aqua v2.2.3 ...
aqua_linux_arm64.tar.gz: OK
===> /tmp/tmp.hlehkM/aqua update-aqua
INFO[0000] download and unarchive the package            aqua_version=2.2.3 env=linux/arm64 new_version=v2.6.0 package_name=aquaproj/aqua package_version=v2.6.0 program=aqua registry=
INFO[0001] verify a package with slsa-verifier           aqua_version=2.2.3 env=linux/arm64 new_version=v2.6.0 package_name=aquaproj/aqua package_version=v2.6.0 program=aqua registry=
INFO[0001] download and unarchive the package            aqua_version=2.2.3 env=linux/arm64 new_version=v2.6.0 package_name=slsa-framework/slsa-verifier package_version=v2.1.0 program=aqua registry=
Verified signature against tlog entry index 20223381 at URL: https://rekor.sigstore.dev/api/v1/log/entries/24296fb24b8ad77a607c980c833eb73f84b6461d7932b893a0cc206bd8289cf74c92137efedf66c6
Verified build using builder https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.5.0 at commit 903d205f6876aba423f753613ff01bbf97216c00
Verifying artifact /tmp/467478560: PASSED

PASSED: Verified SLSA provenance
INFO[0010] create a symbolic link                        aqua_version=2.2.3 command=aqua env=linux/arm64 new_version=v2.6.0 package_name=aquaproj/aqua package_version=v2.6.0 program=aqua
aqua version 2.6.0 (903d205f6876aba423f753613ff01bbf97216c00)
/workspace
INFO[0000] download and unarchive the package            aqua_version=2.6.0 env=linux/arm64 package_name=aqua-proxy package_version=v1.2.0 program=aqua registry=
INFO[0000] create a symbolic link                        aqua_version=2.6.0 command=aqua-proxy env=linux/arm64 package_name=aqua-proxy package_version=v1.2.0 program=aqua registry=
INFO[0001] create a symbolic link                        aqua_version=2.6.0 command=task env=linux/arm64 program=aqua
INFO[0001] create a symbolic link                        aqua_version=2.6.0 command=terraform env=linux/arm64 program=aqua
terraform init
INFO[0000] download and unarchive the package            aqua_version=2.6.0 env=linux/arm64 exe_name=terraform package=hashicorp/terraform package_name=hashicorp/terraform package_version=v1.4.6 program=aqua registry=standard

Initializing the backend...

Initializing provider plugins...
- Finding latest version of hashicorp/null...
- Installing hashicorp/null v3.2.1...
- Installed hashicorp/null v3.2.1 (signed by HashiCorp)

Terraform has created a lock file .terraform.lock.hcl to record the provider
selections it made above. Include this file in your version control repository
so that Terraform can guarantee to make the same selections by default when
you run "terraform init" in the future.

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

</details>

```console
$ task tf-init
```

<details>

```console
$ task tf-init
task: [setup-aqua] bash scripts/setup_aqua.sh
task: [tf-init] terraform init

Initializing the backend...

Initializing provider plugins...
- Reusing previous version of hashicorp/null from the dependency lock file
- Installing hashicorp/null v3.2.1...
- Installed hashicorp/null v3.2.1 (signed by HashiCorp)

Terraform has made some changes to the provider dependency selections recorded
in the .terraform.lock.hcl file. Review those changes and commit them to your
version control system if they represent changes you intended to make.

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

</details>

Directory structure

```
Makefile or Taskfile.yml
scripts/
  setup_aqua.sh
```

setup_aqua.sh

```bash
#!/usr/bin/env bash

set -eu
set -o pipefail

if command -v aqua > /dev/null 2>&1; then
  exit 0
fi

tempdir=$(mktemp -d)
cd "$tempdir"
curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer
echo "98b883756cdd0a6807a8c7623404bfc3bc169275ad9064dc23a6e24ad398f43d  aqua-installer" | sha256sum -c -
chmod +x aqua-installer

./aqua-installer
cd -

rm -R "$tempdir"

aqua i -l
```

Makefile

```makefile
ifeq ($(AQUA_ROOT_DIR),)
ifeq ($(XDG_DATA_HOME),)
	AQUA_ROOT_DIR := $(HOME)/.local/share/aquaproj-aqua
else
	AQUA_ROOT_DIR := $(XDG_DATA_HOME)/aquaproj-aqua
endif
endif

PATH := $(AQUA_ROOT_DIR)/bin:$(PATH)

.PHONY: setup-aqua
setup-aqua:
	bash scripts/setup_aqua.sh

.PHONY: tf-init
tf-init: setup-aqua
	terraform init
```

Taskfile.yml

```yaml
version: '3'

vars:
  AQUA_ROOT_DIR:
    sh: echo "${AQUA_ROOT_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/aquaproj-aqua}"

env:
  PATH: "{{.AQUA_ROOT_DIR}}/bin:{{.PATH}}"

tasks:
  tf-init:
    deps: [setup-aqua]
    cmds:
      - terraform init

  setup-aqua:
    cmds:
      - bash scripts/setup_aqua.sh
```
