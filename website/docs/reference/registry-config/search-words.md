---
sidebar_position: 610
---

# search_words

[#972](https://github.com/aquaproj/aqua/pull/972) aqua >= v1.16.0 is required

`search_words` is used only for searching packages by `aqua g` command.
aqua searches not only the package name, aliases, and command names but also `search_words`.

For example, it is useful to find GitHub CLI with the word `github`.

e.g.

```yaml
packages:
  - type: github_release
    repo_owner: cli
    repo_name: cli
    description: GitHub’s official command line tool
    search_words:
      - github
```

```
  cli/cli [gh]: github
  github/licensed
> github/hub
  8/643
> github
```
