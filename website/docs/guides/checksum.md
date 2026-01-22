---
sidebar_position: 15
---

# Enable Checksum Verification

About Checksum Verification, please see also.

- [Reference](/docs/reference/security/checksum)
- [Configuration](/docs/reference/config/checksum)
- [Registry Configuration](/docs/reference/registry-config/checksum)
- [Usage > aqua update-checksum](/docs/reference/usage#aqua-update-checksum)

## Create a GitHub Repository

[Let's create a GitHub Repository for this tutorial](https://github.com/new).
You can remove the repository after this tutorial.

## Prepare GitHub Access Token

Please create a classic personal access token and add it to Repository Secrets.

- name: GH_TOKEN
- required permissions: `contents: write`

:::caution
GitHub Actions' token `GITHUB_TOKEN` is unavailable.
:::

:::info
~~Unfortunately, fine-grained personal access token is unavailable at the moment because it doesn't support GraphQL API.~~
~~https://github.com/cli/cli/issues/6680~~

2023-04-27 [fine-grained access token supports GraphQL API now.](https://github.blog/changelog/2023-04-27-graphql-improvements-for-fine-grained-pats-and-github-apps/)
:::

:::info
In this time we use a classic personal access token, but we recommend GitHub App or fine-grained access token in terms of security.
:::

## Create aqua.yaml

```sh
aqua init
aqua g -i suzuki-shunsuke/tfcmt
```

## Enable Checksum Verification

By default, checksum verification is disabled.
Let's edit aqua.yaml and enable Checksum Verification.

```yaml
---
checksum:
  enabled: true
registries:
- type: standard
  ref: v4.155.1 # renovate: depName=aquaproj/aqua-registry
packages:
- name: suzuki-shunsuke/tfcmt@v4.2.0
```

## Set up GitHub Actions Workflow

:::caution
For CircleCI Users, please use [circleci-orb-aqua's update-checksum command](https://circleci.com/developer/orbs/orb/aquaproj/aqua#commands-update-checksum) instead.
:::

To create and update `aqua-checksum.json` automatically, let's set up GitHub Actions.

```sh
mkdir -p .github/workflows
vi .github/workflows/update-aqua-checksum.yaml
```

```yaml
name: update-aqua-checksum
on:
  pull_request:
    paths:
      - aqua.yaml
      - aqua-checksums.json
jobs:
  update-aqua-checksums:
    uses: aquaproj/update-checksum-workflow/.github/workflows/update-checksum.yaml@d248abb88efce715d50eb324100d9b29a20f7d18 # v1.0.4
    permissions:
      contents: read
    with:
      aqua_version: v2.48.3
      prune: true
    secrets:
      gh_token: ${{secrets.GH_TOKEN}}
      # gh_app_id: ${{secrets.APP_ID}}
      # gh_app_private_key: ${{secrets.APP_PRIVATE_KEY}}
```

We use [update-checksum-action](https://github.com/aquaproj/update-checksum-action).

## Create a pull request

Commit `aqua.yaml` and `.github/workflows/update-aqua-checksum.yaml`.

```sh
git checkout -b ci/aqua-checksum
git add aqua.yaml .github/workflows/update-aqua-checksum.yaml
git commit -m "ci: add aqua.yaml and set up workflow"
git push origin ci/aqua-checksum
```

Create a pull request. Then `aqua-checksums.json` will be created by GitHub Actions.

![image](https://user-images.githubusercontent.com/13323303/224527388-720ce451-bdce-4055-9eed-ba0942615eea.png)

![image](https://user-images.githubusercontent.com/13323303/224527533-8fc150e2-55c1-4ca4-a9c7-f05544fdeccb.png)

## Change a package version

Let's change version.

```sh
sed -i "s/v4.2.0/v4.1.0/" aqua.yaml
```

```diff
-- name: suzuki-shunsuke/tfcmt@v4.2.0
+- name: suzuki-shunsuke/tfcmt@v4.1.0
```

Push a commit.

```sh
git pull origin ci/aqua-checksum
git add aqua.yaml
git commit -m "chore: change tfcmt version"
git push origin "ci/aqua-checksum"
```

Then `aqua-checksums.json` is updated automatically.

![image](https://user-images.githubusercontent.com/13323303/224527976-4ddb1607-9958-4269-8882-3c0657e98a72.png)

![image](https://user-images.githubusercontent.com/13323303/224528023-72aba252-7507-47fa-87b2-dc08eb7f908b.png)

## See how Checksum Verification prevents tampering

Let's see how Checksum Verification prevents tampering.
It's bothersome to tamper assets actually, so in this time let's simulate the situation by tampering checksum in `aqua-checksums.json`.

```sh
git pull origin ci/aqua-checksum
vi aqua-checksums.json
```

```diff
     {
       "id": "github_release/github.com/suzuki-shunsuke/tfcmt/v4.1.0/tfcmt_linux_amd64.tar.gz",
-      "checksum": "A8E55BEA1A5F94F9515FD9C5C3296D1874461BA1DBD158B3FC0ED6A0DB3B7D91",
+      "checksum": "A8E55BEA1A5F94F9515FD9C5C3296D1874461BA1DBD158B3FC0ED6A0DB3B7D92",
       "algorithm": "sha256"
     },
```

Add a GitHub Actions job that runs a tampered package.

```yaml
  test:
    runs-on: ubuntu-24.04
    permissions:
      contents: read
    env:
      AQUA_LOG_COLOR: always
      AQUA_REQUIRE_CHECKSUM: "true"
    steps:
      - uses: actions/checkout@71cf2267d89c5cb81562390fa70a37fa40b1305e # v6-beta
      - uses: aquaproj/aqua-installer@11dd79b4e498d471a9385aa9fb7f62bb5f52a73c # v4.0.4
        with:
          aqua_version: v2.48.3
        env:
          GITHUB_TOKEN: ${{github.token}}
      - run: tfcmt -v
```

```sh
git add aqua-checksums.json
git commit -m "chore: tamper aqua-checksums.json"
git push origin "ci/aqua-checksum"
```

Then `test` job would fail because the checksum is unmatched.

![image](https://user-images.githubusercontent.com/13323303/224528789-eeda95e7-73b9-46a3-95da-da954087e83b.png)

```
time="2023-03-12T06:36:05Z" level=fatal msg="aqua failed" actual_checksum=A8E55BEA1A5F94F9515FD9C5C3296D1874461BA1DBD158B3FC0ED6A0DB3B7D91 aqua_version=2.28.0 env=linux/amd64 error="checksum is invalid" exe_name=tfcmt expected_checksum=A8E55BEA1A5F94F9515FD9C5C3296D1874461BA1DBD158B3FC0ED6A0DB3B7D92 package=suzuki-shunsuke/tfcmt package_version=v4.1.0 program=aqua
```

## Please consider autofix.ci or Securefix Action instead of update-checksum-action and update-checksum-workflow

Instead of [update-checksum-action](https://github.com/aquaproj/update-checksum-action) and [update-checksum-workflow](https://github.com/aquaproj/update-checksum-workflow), we recommend [autofix.ci](https://autofix.ci/) or [Securefix Action](https://github.com/securefix-action/action) for security.

- autofix.ci: For OSS
- Securefix Action: For private repositories

### autofix.ci

About autofix.ci, please see the website. https://autofix.ci/
autofix.ci is free for OSS.
autofix.ci has various benefits:

- You can fix pull requests from fork securely
- Easy to use. You don't need to take care of how to create and push commits
- Commits are verified
- You no longer need to branch processing based on whether the pull request is from a fork or not

We're using autofix.ci in various places.

e.g. https://github.com/aquaproj/aqua-renovate-config/blob/main/.github/workflows/autofix.yaml

This is an example workflow:

```yaml
name: autofix.ci
on: pull_request
permissions: {}
jobs:
  autofix:
    runs-on: ubuntu-24.04
    permissions: {}
    timeout-minutes: 15
    steps:
      - name: Checkout the repository
        uses: actions/checkout@71cf2267d89c5cb81562390fa70a37fa40b1305e # v6-beta
        with:
          persist-credentials: false
      - name: Install aqua
        uses: aquaproj/aqua-installer@11dd79b4e498d471a9385aa9fb7f62bb5f52a73c # v4.0.4
        with:
          aqua_version: v2.43.0
      - name: Fix aqua-checksums.json
        run: aqua upc -prune
      - name: Run autofix.ci
        uses: autofix-ci/action@635ffb0c9798bd160680f18fd73371e355b85f27 # v1.3.2
```

### Securefix Action

[About Securefix Action, please see the document.](https://github.com/securefix-action/action)
You can update aqua-checksums.json using Securefix Action and `aqua upc` command:

e.g.

```yaml
name: Update aqua-checksums.json
on: pull_request
permissions: {}
jobs:
  securefix:
    runs-on: ubuntu-24.04
    permissions: {}
    timeout-minutes: 15
    steps:
      - name: Checkout the repository
        uses: actions/checkout@71cf2267d89c5cb81562390fa70a37fa40b1305e # v6-beta
        with:
          persist-credentials: false
      - name: Install aqua
        uses: aquaproj/aqua-installer@11dd79b4e498d471a9385aa9fb7f62bb5f52a73c # v4.0.4
        with:
          aqua_version: v2.43.0
      - name: Fix aqua-checksums.json
        run: aqua upc -prune
      - name: Commit and push
        uses: securefix-action/action@b6d50c16ddf4b0d137e42ad4fa0ee29dc43d4b55 # v0.5.4
        with:
          app_id: ${{secrets.APP_ID}}
          app_private_key: ${{secrets.APP_PRIVATE_KEY}}
          server_repository: demo-client
```

### commit-action

You can also use [suzuki-shunsuke/commit-action](https://github.com/suzuki-shunsuke/commit-action).
But we recommend Securefix Action for security.

e.g.

```yaml
name: Update aqua-checksums.json
on: pull_request
permissions: {}
jobs:
  update-aqua-checksums:
    runs-on: ubuntu-24.04
    permissions: {}
    timeout-minutes: 15
    steps:
      - name: Checkout the repository
        uses: actions/checkout@71cf2267d89c5cb81562390fa70a37fa40b1305e # v6-beta
        with:
          persist-credentials: false
      - name: Install aqua
        uses: aquaproj/aqua-installer@11dd79b4e498d471a9385aa9fb7f62bb5f52a73c # v4.0.4
        with:
          aqua_version: v2.43.0
      - name: Fix aqua-checksums.json
        run: aqua upc -prune
      - name: Commit and push
        uses: suzuki-shunsuke/commit-action@f28421acc277a6d6a9c1f94ea449076ad77dba67 # v0.1.0
        with:
          app_id: ${{secrets.APP_ID}}
          app_private_key: ${{secrets.APP_PRIVATE_KEY}}
```
