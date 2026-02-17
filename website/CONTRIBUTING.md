# Contributing

## Set up

Requirements:

- Node.js
- aqua

```sh
aqua i -l
```

### Commit Signing

https://github.com/suzuki-shunsuke/oss-contribution-guide/blob/main/docs/commit-signing.md

## Run the document

```sh
npm i
npm start
```

## Check typo using typos

https://github.com/crate-ci/typos

```sh
typos -w .
```

If the fix by typos is wrong, please fix [_typos.toml](/_typos.toml).

## Update Usage

1. Update aqua to the latest version.
2. Run `cmdx usage`

```sh
aqua upa
cmdx usage
```
