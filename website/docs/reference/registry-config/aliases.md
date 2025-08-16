---
sidebar_position: 600
---

# aliases

[#674](https://github.com/aquaproj/aqua/pull/674) [#675](https://github.com/aquaproj/aqua/pull/675) aqua >= v1.5.0 is required

Aliases of packages.

e.g.

```yaml
packages:
- name: ahmetb/kubectx/kubens
  aliases:
  - name: ahmetb/kubens
```

You can use the alias as the package name in `aqua.yaml`, and alias can be used in `aqua g` command.

`aliases` is used to keep the compatibility when the package name is changed.
Sometimes the package name is changed because the repository is renamed or transferred.

## Use `aliases` only for keeping the compatibility

:::caution
`aliases` should be used only for keeping the compatibility.
:::

https://github.com/aquaproj/aqua-registry/pull/4538#discussion_r911871799

> I think the same package should have only one name, and aliases should be used only to keep the compatibility.
> Users may be confused what's the difference of kubernetes-sigs/controller-tools/controller-gen and kubernetes-sigs/kubebuilder/controller-gen and which package they should use.
> 
> On the other hand, it is useful to allow to search packages with additional words.
> To do this, I'll add a new field [search_words](search-words.md).
> 
> e.g.
> 
> ```yaml
> name: kubernetes-sigs/controller-tools/controller-gen
> search_words:
>   - kubebuilder
> ```
