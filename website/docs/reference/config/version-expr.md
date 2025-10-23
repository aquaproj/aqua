---
sidebar_position: 900
---

# `version_expr`, `version_expr_prefix`

`aqua >= v2.40.0` [#3363](https://github.com/aquaproj/aqua/pull/3363)

`version_expr` and `version_expr_prefix` are fields to read a package version from external files.
For example, you can get a version of Terraform from `.terraform-version`.

.terraform-version:

```
1.10.2
```

```yaml
packages:
- name: hashicorp/terraform
  # The version is `v1.10.2`.
  version_expr: readFile(".terraform-version")
  version_expr_prefix: v
```

You can also read JSON and YAML files:

```yaml
packages:
- name: hashicorp/terraform
  version_expr: readJSON("version.json").version
```

```yaml
packages:
- name: hashicorp/terraform
  version_expr: readYAML("version.yaml").version
```

`version_expr` is evaluated by [expr](https://expr-lang.org/docs/language-definition).
The package version is `version_expr_prefix` + `The evaluation result of version_expr`.

The following custom functions are available:

- `readFile("file path")`: read a file and returns a content
- `readJSON("file path")`: read a JSON file and returns a content 
- `readYAML("file path")`: read a YAML file and returns a content 

To prevent secrets from being leaked by reading secret files via read{File,JSON,YAML} functions, the evaluation result of `version_expr` must match with the regular expression `^v?\d+\.\d+(\.\d+)*[.-]?((alpha|beta|dev|rc)[.-]?)?\d*`.

If the version has a prefix such as `cli-`, the version doesn't match with the regular expression.
In that case, you can set the prefix with `version_expr_prefix`.
