---
sidebar_position: 700
---

# Build Container (Docker) Image

When building a container (Docker) image, you may want to download and install tools from GitHub Releases or other sources.
In particular, if you run CI with CircleCI or Google Cloud Build, you may want to install tools for CI on the image.

Traditionally, you would use curl, tar, unzip, etc. to install these tools, but with aqua, you can declaratively manage them.
You don't have to look up download URLs, formats, etc. yourself.
You can also use Renovate to automate updates.

aqua.yaml

```yaml
---
# aqua - Declarative CLI Version Manager
# https://aquaproj.github.io/
registries:
- type: standard
  ref: v4.155.1  # renovate: depName=aquaproj/aqua-registry
packages:
- name: rhysd/actionlint@v1.6.15
- name: golangci/golangci-lint@v1.47.2
- name: reviewdog/reviewdog@v0.14.1
```

Dockerfile

```dockerfile
FROM debian:bookworm-20240408
ENV PATH=/root/.local/share/aquaproj-aqua/bin:$PATH
COPY aqua.yaml aqua-checksums.json /etc/aqua/
ENV AQUA_GLOBAL_CONFIG=/etc/aqua/aqua.yaml
RUN apt-get update && \
    apt-get install -y curl && \
    curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer && \
    echo "98b883756cdd0a6807a8c7623404bfc3bc169275ad9064dc23a6e24ad398f43d  aqua-installer" | sha256sum -c - && \
    chmod +x aqua-installer && \
    ./aqua-installer -v v2.48.3 && \
    rm aqua-installer && \
    aqua i -a && \
    apt-get remove -y curl && \
    apt-get clean
```

## Remove aqua from Image

aqua >= [v1.18.0](https://github.com/aquaproj/aqua/releases/tag/v1.18.0)

In the above Docker image, aqua is installed and used internally to execute tools.
However, if you do not want to install anything extra in the Docker image, if you want to keep the image minimal, or if you do not want to switch tool versions with aqua,
you can also remove aqua using the Multistage Build and `aqua cp` command.

Dockerfile

```dockerfile
FROM debian:bookworm-20240408 AS aqua
ENV PATH=/root/.local/share/aquaproj-aqua/bin:$PATH
COPY aqua.yaml aqua-checksums.json /etc/aqua/
ENV AQUA_GLOBAL_CONFIG=/etc/aqua/aqua.yaml
RUN apt-get update
RUN apt-get install -y curl
RUN curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.2/aqua-installer
RUN echo "98b883756cdd0a6807a8c7623404bfc3bc169275ad9064dc23a6e24ad398f43d  aqua-installer" | sha256sum -c -
RUN chmod +x aqua-installer
RUN ./aqua-installer -v v2.48.3
RUN aqua i -a
RUN aqua cp -o /dist actionlint reviewdog

FROM debian:bookworm-20240408
COPY --from=aqua /dist/* /usr/local/bin/
```

`aqua cp` installs specified tools and copies executable files to the specified directory.
In the above example, actionlint and reviewdog are installed and copied under /dist.
Only executable files are installed in the final image.

## Notes of `aqua cp`

There is a caveat to `aqua cp`.
`aqua cp` copies only executable files from packages.
Therefore, tools that do not work with a single file will not work properly even if they are copied by `aqua cp`.
If the tool is a single binary written in Go, there is basically no problem, but if it is a shell script depending on another files in the same repository, it will not work properly.

For example, tfenv will not work correctly even if you copy it by `aqua cp`.
You need to install those tools in a different way.
