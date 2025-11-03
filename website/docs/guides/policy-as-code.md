---
sidebar_position: 20
---

# Policy as Code

`aqua >= v2.3.0`

Policy is a feature to restrict the package installation and execution.
The main purpose of Policy is to improve the security by preventing malicious tools from being executed.

If you use only Standard Registry, you don't have to care of Policy.
From aqua v2, aqua allows only Standard Registry by default.
This means aqua prevents malicious tools from being executed via malicious Registries by default.

If you use non Standard Registries, you have to create a Policy file to allow them.

## Getting Started

1. [Set up the environment with Docker](/docs/tutorial/#docker)
1. Create `aqua.yaml` and a local Registry `registry.yaml`
1. Try to use a local Registry and confirm the default Policy
1. Create a Git repository and aqua-policy.yaml
1. Confirm the warning
1. Run `aqua policy deny`
1. Run `aqua policy allow`

--

2. Create `aqua.yaml` and a local Registry `registry.yaml`

```
aqua init
aqua gr suzuki-shunsuke/ci-info > registry.yaml
vi aqua.yaml
aqua g -i cli/cli local,suzuki-shunsuke/ci-info
```

aqua.yaml

```yaml
---
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
# checksum:
#   enabled: true
#   require_checksum: true
#   supported_envs:
#   - all
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry
- name: local
  type: local
  path: registry.yaml
packages:
- name: cli/cli@v2.25.1
- name: suzuki-shunsuke/ci-info@v2.1.2
  registry: local
```

3. Try to use a local Registry and confirm the default Policy

```console
fe179d7889fd:~/workspace$ aqua i
INFO[0000] download and unarchive the package            aqua_version= env=linux/arm64 package_name=aqua-proxy package_version=v1.1.4 program=aqua registry=
INFO[0000] create a symbolic link                        aqua_version= command=aqua-proxy env=linux/arm64 package_name=aqua-proxy package_version=v1.1.4 program=aqua registry=
INFO[0000] create a symbolic link                        aqua_version= command=gh env=linux/arm64 program=aqua
INFO[0000] create a symbolic link                        aqua_version= command=ci-info env=linux/arm64 program=aqua
ERRO[0000] install the package                           aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/002" env=linux/arm64 error="this package isn't allowed" package_name=suzuki-shunsuke/ci-info package_version=v2.1.2 program=aqua registry=local
INFO[0000] download and unarchive the package            aqua_version= env=linux/arm64 package_name=cli/cli package_version=v2.27.0 program=aqua registry=standard
FATA[0002] aqua failed                                   aqua_version= env=linux/arm64 error="it failed to install some packages" program=aqua
```

It fails to install `suzuki-shunsuke/ci-info` because the local Registry isn't allowed by default.

```
ERRO[0000] install the package                           aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/002" env=linux/arm64 error="this package isn't allowed" package_name=suzuki-shunsuke/ci-info package_version=v2.1.2 program=aqua registry=local
```

On the other hand, GitHub CLI is installed properly because Standard Registry is allowed by default.

```console
a82023e65a9e:~/workspace$ gh version
gh version 2.27.0 (2023-04-07)
https://github.com/cli/cli/releases/tag/v2.27.0
```

4. Create a Git repository and aqua-policy.yaml

Let's create a Policy to allow the local Registry.

`.git` directory is required so that aqua finds a Policy file.

```sh
git init # Create .git
aqua policy init
vi aqua-policy.yaml
```

aqua-policy.yaml

```yaml
---
# aqua Policy
# https://aquaproj.github.io/
registries:
# Example
  - name: local
    type: local
    path: registry.yaml
# - name: aqua-registry
#   type: github_content
#   repo_owner: aquaproj
#   repo_name: aqua-registry
#   ref: semver(">= 3.0.0") # ref is optional
#   path: registry.yaml
  - type: standard
    ref: semver(">= 3.0.0")
packages:
# Example
  - registry: local # allow all packages in the Registry
# - name: cli/cli # allow only a specific package. The default value of registry is "standard"
# - name: cli/cli
#   version: semver(">= 2.0.0") # version is optional
  - registry: standard
```

5. Confirm the warning

Run `aqua i`, then aqua outputs the warning and it fails to install `suzuki-shunsuke/ci-info`.

```console
fe179d7889fd:~/workspace$ aqua i
WARN[0000] The policy file is ignored unless it is allowed by "aqua policy allow" command.

$ aqua policy allow "/home/foo/workspace/aqua-policy.yaml"

If you want to keep ignoring the policy file without the warning, please run "aqua policy deny" command.

$ aqua policy deny "/home/foo/workspace/aqua-policy.yaml"

   aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/003" env=linux/arm64 policy_file=/home/foo/workspace/aqua-policy.yaml program=aqua
ERRO[0000] install the package                           aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/002" env=linux/arm64 error="this package isn't allowed" package_name=suzuki-shunsuke/ci-info package_version=v2.1.2 program=aqua registry=local
FATA[0000] aqua failed                                   aqua_version= env=linux/arm64 error="it failed to install some packages" program=aqua
```

To resolve the warning, you have to check the Policy file and run either `aqua policy allow` or `aqua policy deny`.
If the Policy file is reliable, please run `aqua policy allow`.

6. Run `aqua policy deny`

Before running `aqua policy allow`, let's try to run `aqua policy deny`.

```
aqua policy deny "/home/foo/workspace/aqua-policy.yaml"
```

ci-info still failed but the warning is suppressed.

```console
2f4a758ab4ef:~/workspace$ ci-info --help
FATA[0000] aqua failed                                   aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/002" env=linux/arm64 error="this package isn't allowed" exe_name=ci-info package=suzuki-shunsuke/ci-info package_version=v2.1.2 program=aqua
```

7. Run `aqua policy allow`.

```
aqua policy allow "/home/foo/workspace/aqua-policy.yaml"
```

Then ci-info is available.

```console
2f4a758ab4ef:~/workspace$ ci-info --version
INFO[0000] download and unarchive the package            aqua_version= env=linux/arm64 exe_name=ci-info package=suzuki-shunsuke/ci-info package_name=suzuki-shunsuke/ci-info package_version=v2.1.2 program=aqua registry=local
ci-info version 2.1.2 (4a047e648dd0b9d0de1be356421d5d043c38d080)
```

If you modify the Policy file, you have to allow the change again.

```
echo "" >> aqua-policy.yaml
```

```console
2f4a758ab4ef:~/workspace$ ci-info --version
WARN[0000] The policy file is changed. The policy file is ignored unless it is allowed by "aqua policy allow" command.

$ aqua policy allow "/home/foo/workspace/aqua-policy.yaml"

If you want to keep ignoring the policy file without the warning, please run "aqua policy deny" command.

$ aqua policy deny "/home/foo/workspace/aqua-policy.yaml"

   aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/003" env=linux/arm64 exe_name=ci-info package=suzuki-shunsuke/ci-info package_version=v2.1.2 policy_file=/home/foo/workspace/aqua-policy.yaml program=aqua
FATA[0000] aqua failed                                   aqua_version= doc="https://aquaproj.github.io/docs/reference/codes/002" env=linux/arm64 error="this package isn't allowed" exe_name=ci-info package=suzuki-shunsuke/ci-info package_version=v2.1.2 program=aqua
```

Please run `aqua policy allow` again, then ci-info is available.

```console
2f4a758ab4ef:~/workspace$ aqua policy allow "/home/foo/workspace/aqua-policy.yaml"
2f4a758ab4ef:~/workspace$ ci-info --version
ci-info version 2.1.2 (4a047e648dd0b9d0de1be356421d5d043c38d080)
```

Basically Policy files aren't changed so frequently, so it wouldn't be so bothersome to run `aqua policy allow`.

## aqua-installer's `policy_allow` input

aqua >= `v2.3.0`, aqua-installer >= `v2.1.0`

If the input `policy_allow` is set, aqua-installer runs `aqua policy allow` command.

```yaml
- uses: aquaproj/aqua-installer@11dd79b4e498d471a9385aa9fb7f62bb5f52a73c # v4.0.4
  with:
    aqua_version: v2.48.3
    policy_allow: "true"
```

## update-checksum-workflow's `policy_allow` input

aqua >= `v2.3.0`, [update-checksum-workflow](/docs/products/update-checksum-workflow) >= `v0.1.5`

If the input `policy_allow` is set, `aqua policy allow` is run.

## :bulb: Best practice: Configure CODEOWENRS to protect Policy files

Basically you don't have to change Policy files so frequently and the change of Policy files should be reviewed carefully in terms of security.
So it is a good practice to protect Policy files by CODEOWNERS.

## See also

- [Reference](/docs/reference/security/policy-as-code)
