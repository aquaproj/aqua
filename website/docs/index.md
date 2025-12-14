---
sidebar_position: 100
---

# Home

## :bulb: NotebookLM And DeepWiki

- [Google NotebookLM](https://notebooklm.google.com/notebook/874e89e4-66a1-459a-82c9-923b81501a71)
- [DeepWiki](https://deepwiki.com/aquaproj/aqua)

## How to learn aqua

aqua has a lot of documents, so it's hard to read all of them.
But don't worry. You don't have to read all of them all at once.

Here is a brief overview of how to learn aqua.

1. To install aqua, please read [Install](/docs/install)
1. To learn the basic usage, please try [Tutorial](/docs/tutorial). aqua is easy to use, so you would be able to use aqua in a short time
1. To learn the overview, please see [Demo](https://asciinema.org/a/498262?autoplay=1), [top page](/), and this page ([Introduction](#introduction), [Why aqua](#why-aqua), [Comparison](#comparison))
1. To learn advanced usage, please read [Guides](/docs/guides)
1. To contribute to Standard Registry (Add new packages, fix bugs), please read [Contributing](/docs/products/aqua-registry/contributing)
1. To develop custom Registry, please read [Develop a Registry](/docs/develop-registry/)

## Contact us

If you have any question, please contact us.

- [GitHub Issues](https://github.com/aquaproj/aqua/issues/new/choose)
- X (Formerly Twitter):
  - English: [@aquaclivm](https://x.com/aquaclivm), [@szkdash_en](https://x.com/szkdash_en)
  - Japanese: [@szkdash](https://x.com/szkdash)
- Slack
  - [English](https://gophers.slack.com/archives/C04RALTG29K), [Japanese](https://gophers.slack.com/archives/C04RDHZQ9K5)
  - [Please request to join Gophers Workspace here.](https://github.com/aquaproj/aquaproj.github.io/issues/1245)
- [Discord](https://discord.gg/CBGesz8yBX)

[If you have any trouble in joining our Slack or Discord, please create an issue.](https://github.com/aquaproj/aquaproj.github.io/issues/new?assignees=&labels=community-join-issue&projects=&template=community-join-problem.yml)

## For Japanese

We published an e-book about aqua for free. Please read it!

[aqua CLI Version Manager 入門 (2023-10-01)](https://zenn.dev/shunsuke_suzuki/books/aqua-handbook)

## Introduction

aqua is a declarative CLI Version Manager written in Go.
You can manage tool versions with YAML.

e.g. aqua.yaml

```yaml
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry

packages:
- name: cli/cli@v2.2.0
- name: junegunn/fzf@0.28.0
```

The short demo would be useful to understand aqua.

[![asciicast](https://asciinema.org/a/498262.svg)](https://asciinema.org/a/498262?autoplay=1)

You can install tools simply by `aqua i` command.

```console
$ aqua i
```

aqua supports various tools officially.
You can search tools interactively by `aqua g` command.

```console
$ aqua g
```

```console
  newrelic/newrelic-cli (standard) (newrelic)                   ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
  pivotal-cf/pivnet-cli (standard) (pivnet)                     │  cli/cli
  scaleway/scaleway-cli (standard) (scw)                        │
  tfmigrator/cli (standard) (tfmigrator)                        │  https://cli.github.com/
  aws/copilot-cli (standard) (copilot)                          │  GitHub’cs official command line tool
  codeclimate/test-reporter (standard)                          │
  create-go-app/cli (standard) (cgapp)                          │
  harness/drone-cli (standard) (drone)                          │
  sigstore/rekor (standard) (rekor-cli)                         │
  getsentry/sentry-cli (standard)                               │
  grafana/loki/logcli (standard)                                │
  knative/client (standard) (kn)                                │
  rancher/cli (standard) (rancher)                              │
  tektoncd/cli (standard) (tkn)                                 │
  civo/cli (standard) (civo)                                    │
  dapr/cli (standard) (dapr)                                    │
  mongodb/mongocli (standard)                                   │
  openfaas/faas-cli (standard)                                  │
> cli/cli (standard) (gh)                                       │
  50/399                                                        │
> cli                                                           └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
```

To add supported tools, please see [here](/docs/products/aqua-registry/contributing).

## Usecase

1. [Install tools in your laptop globally](tutorial/global-config.md). You can also manage tools in your `dotfiles` repository
1. Manage tools for projects
1. [Manage tools for your organization and team](guides/team-config.md)
1. [Distribute private tools in your organization](guides/private-package.md)

## Why aqua?

- Change tool versions per project
  - Unlike Package Manager such as Homebrew, aqua supports changing tool version per project
- Unify tool versions to prevent problems due to version difference
  - aqua makes you specify tool versions strictly
  - aqua supports cross platforms and provides the unified way to manage tools both in local development and CI
- [Easy to use](#easy-to-use)
  - This is so important for introducing a tool to a project and having developers use it
- Painless
  - aqua installs tools automatically when they are triggered. You don't have to run `aqua i` everytime tools are updated
  - [Continuous update by Renovate](guides/renovate.md)
  - Easy to support new tools. You don't have to maintain plugins or something yourself. You only have to send a pull request to [the Standard Registry](https://github.com/aquaproj/aqua-registry), which is very easy
  - 👋 [Good bye shell script](#-good-bye-shell-script)
    - You don't have to write similar shell scripts to install tools many times. You only have to manage tools declaratively with YAML and run `aqua i`
- 🛡️ [Security](reference/security/index.md)
  - aqua supports security features such as [Checksum Verification](guides/checksum.md), [Policy as Code](/docs/reference/security/policy-as-code), [Cosign and SLSA Provenance Support](reference/security/cosign-slsa.md), and [Minisign](reference/security/minisign.md)
- Lower overhead than container
  - You don't have to suffer from container specific issues
- [Support private packages](guides/private-package.md)

### Strict Version Control

The difference of tool version often causes troubles.
In the team development, if developers use different tool versions each others, you would suffer from troubles due to the version difference.
And if the tool version in your laptop is different from the tool version in CI, you would suffer from the same trouble.

![image](https://user-images.githubusercontent.com/13323303/221387356-95d9d9fd-6c19-4015-830c-63d86a4a38e0.png)

And if multiple projects require the different tool versions, you have to switch the tool version per project somehow.

Package managers such as Homebrew and apt don't support switching tool versions.

aqua forces to pin tool versions and supports switching tool versions per project.
You can upgrade a tool temporaroly and rollback the upgrade easily only by editing `aqua.yaml`.

Of course, aqua supports Monorepo which uses different tool versions per service as well.

#### Installing `latest` version in CI is danger

If you install a tool in CI, the tool should be version controlled to avoid troubles.
Installing `latest` version is useful but danger because the `latest` version is changed suddenly and unexpected trouble may occur.
Your CI would be broken suddenly though you don't change any code.
Many GitHub Actions install `latest` version by default.

aqua forces to pin tool versions and doesn't allow to install `latest` version.

### Easy to install tools for your projects

You can list up tools and their versions for your project in `aqua.yaml`.
Developers can install tools easily only by running `aqua i [-l]`.
You don't have to write the document about which tools are required and how to install them.

### 👋 Good bye shell script

You may install tools by shell scripts.

e.g.

```sh
#!/usr/bin/env bash

set -eu

tempdir=$(mktemp -d)
cd "$tempdir"
curl -Lq -O https://github.com/suzuki-shunsuke/tfcmt/releases/download/v4.2.0/tfcmt_darwin_amd64.tar.gz
tar xvzf tfcmt_darwin_amd64.tar.gz
chmod a+x tfcmt
sudo mv tfcmt /usr/local/bin
rm -R "$tempdir"
```

Shell scripts have some issues.

- You have to maintain scripts per tool
- You have to update the tool version somehow
  - In many cases, the tool isn't updated and the old version is used for a long time
- The script isn't portable
  - The above script supports only darwin/amd64
- The script doesn't verify the checksum
  - If you verify the checksum in shell script, you have to update the checksum along with the version. It's bothersome

Using aqua, you don't have to maintain shell scripts anymore.
aqua supports updating tools by Renovate and checksum verification. aqua works on cross platform.
aqua supports updating not only versions but also checksums automatically.

### Continuous update by Renovate

If tools are version controlled, they should be updated continuously.
Otherwise, they would become old soon, which causes several issues.

aqua provides [Renovate Config Preset](https://github.com/aquaproj/aqua-renovate-config) for continuous update by Renovate.
Using this preset, you can easily update tools by Renovate.

[ref. Update packages by Renovate](/docs/guides/renovate)

:::info
[As of aqua v2.14.0, aqua also supports `update` command.](/docs/guides/update-command)
:::

### 🛡️ Security

You should verify the checksum of the tool before installing the tool.
Otherwise, the tool may be tampered and the malicious code may be executed.

Unfortunately, many shell scripts, asdf plugins, and GitHub Actions don't verify the checksum.

On the other hand, aqua supports the checksum verification.

[ref. Security](reference/security/index.md)

### Easy to use

If you use tools by yourself, you can use any tools you like freely even if the tool is difficult and the learning cost is high.
But when you introduce a tool to your team and organization, it is important that the tool is easy to use.

Even if you have a high motivation to learn them, other members don't necessarily  have the motivation.
They have to focus on their own tasks and don't have a time to learn new tools.
Then you wouldn't be able to introduce the tool to your team well and even if you introduce it some of other members wouldn't use it.

Compared with alternatives such as `asdf` and `tea`, aqua is much easy to use.
Other members have to do only the following things.

1. [Install aqua in their laptops once](/docs/install)
1. Run `aqua i -l`

aqua provides various features, but other members can use aqua without learning them.

### Easy to support new tools

aqua has [the central Registry (Standard Registry)](https://github.com/aquaproj/aqua-registry) and you can add new tools to the Registry.
You don't have to maintain plugins or something yourself. You only have to create issues or pull requests.
To send a pull request you have to write Registry Configuration, but aqua provides [the tool](https://github.com/aquaproj/registry-tool) to scaffold Registry Configuration and create pull requests, so you can send pull requests easily.
Registry Configuration is a declarative YAML files, so you don't have to write shell scripts or something.
Declarative YAML files are much easier to maintain than scripts.

[Many contributors have already contributed to Standard Registry](https://github.com/aquaproj/aqua-registry/graphs/contributors).
Your contribution is welcome!

## Comparison

:::caution
We are not necessarily familiar with compared tools.
So maybe the description about them is wrong.
Your contribution is welcome.
:::

### Compared with Homebrew

- :thumbsup: [Strict Version Control](#strict-version-control)
- :thumbsup: [Windows Support](/docs/reference/windows-support/)

You can use Homebrew to install tools aqua can't install.

### Compared with asdf

- :thumbsup: [Easy to use](#easy-to-use)
- :thumbsup: [Lazy Install](https://aquaproj.github.io/docs/tutorial/lazy-install/)
- :thumbsup: You don't have to install plugins in advance
- :thumbsup: [Continuous update by Renovate](#continuous-update-by-renovate)
- :thumbsup: [Security](reference/security/index.md) ([Checksum Verification](/docs/guides/checksum/))
- :thumbsup: [aqua doesn't force to manage a tool by aqua in a project even if aqua is used to manage the tool in the other project](#aqua-doesnt-force-to-manage-a-tool-by-aqua-in-a-project-even-if-aqua-is-used-to-manage-the-tool-in-the-other-project)
- :thumbsup: [aqua Registry is much easier to maintain than asdf plugins](#easy-to-support-new-tools)
- :thumbsup: Small things
  - :thumbsup: [Share aqua configuration for teams and organizations with AQUA_GLOBAL_CONFIG](/docs/guides/team-config)
  * :thumbsup: [Split the list of packages](/docs/guides/split-config)

#### aqua doesn't force to manage a tool by aqua in a project even if aqua is used to manage the tool in the other project

For example, you can't use both asdf and nvm to manage Node.js in the same laptop.
If you develop a project A using asdf to manage Node.js in your laptop, you can't develop a project B using nvm in the same laptop.

On the other hand, aqua can be used along with alternatives in the same laptop.

Please see [Use aqua combined with other version manager such as asdf](/docs/reference/use-aqua-with-other-tools).

#### Use aqua along with asdf

asdf supports language runtimes such as Python, Ruby, and so on, though aqua can't support them.
You can use asdf for them and can use aqua for other tools.

:::info
2024-08-24 [aqua has supported Node.js :tada:](/docs/reference/nodejs-support)
:::

### Compared with GitHub Actions

- :thumbsup: [Strict Version Control](#strict-version-control)
- :thumbsup: Unify how to install tools in both local development and CI
- :thumbsup: [Continuous update by Renovate](#continuous-update-by-renovate)
- :thumbsup: [Security](reference/security/index.md) ([Checksum Verification](/docs/guides/checksum/))

## Restriction

Please see [Restriction](/docs/reference/restriction/).

### Twitter

We share news about aqua using a Twitter Account [@aquaclivm](https://twitter.com/aquaclivm).
We check tweets about aqua, but it is difficult to search tweets about aqua with the keyword "aqua" because aqua is a very common word.
So when you tweet about aqua, please mention @aquaclivm or use the hash tag [#aquaclivm](https://twitter.com/hashtag/aquaclivm).

## GitHub Sponsor

We'll really appreciate if you become a sponsor of this project!

https://github.com/sponsors/aquaproj
