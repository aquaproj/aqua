# This Dockerfile is used by `cmdx docker`.
FROM mirror.gcr.io/ubuntu:24.04@sha256:1e622c5f073b4f6bfad6632f2616c7f59ef256e96fe78bf6a595d1dc4376ac02 AS base
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update
RUN apt-get install -y sudo vim ca-certificates
RUN echo 'foo ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers
RUN useradd -u 900 -m -r foo
USER foo
ENV PATH=/home/foo/.local/share/aquaproj-aqua/bin:$PATH
RUN mkdir /home/foo/workspace
WORKDIR /home/foo/workspace

FROM base AS build
COPY dist/aqua-docker /usr/local/bin/aqua

FROM base AS prebuilt
RUN sudo apt-get install -y curl
RUN curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v3.1.2/aqua-installer
RUN echo "9a5afb16da7191fbbc0c0240a67e79eecb0f765697ace74c70421377c99f0423  aqua-installer" | sha256sum -c -
RUN chmod +x aqua-installer
RUN ./aqua-installer
