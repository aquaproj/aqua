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

FROM mirror.gcr.io/alpine:3.23.4 AS alpine-base
RUN apk add sudo vim ca-certificates bash
RUN echo '%wheel ALL=(ALL:ALL) NOPASSWD: ALL' >> /etc/sudoers
RUN adduser -u 900 -G wheel -D foo
USER foo
ENV PATH=/home/foo/.local/share/aquaproj-aqua/bin:$PATH
RUN mkdir /home/foo/workspace
WORKDIR /home/foo/workspace

FROM base AS build
COPY dist/aqua-docker /usr/local/bin/aqua

FROM alpine-base AS alpine-build
COPY dist/aqua-docker /usr/local/bin/aqua

FROM base AS prebuilt
RUN sudo apt-get install -y curl
RUN curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.4/aqua-installer
RUN echo "acd21cbb06609dd9a701b0032ba4c21fa37b0e3b5cc4c9d721cc02f25ea33a28  aqua-installer" | sha256sum -c -
RUN chmod +x aqua-installer
RUN ./aqua-installer

FROM alpine-base AS alpine-prebuilt
RUN sudo apk add curl
RUN curl -sSfL -O https://raw.githubusercontent.com/aquaproj/aqua-installer/v4.0.4/aqua-installer
RUN echo "acd21cbb06609dd9a701b0032ba4c21fa37b0e3b5cc4c9d721cc02f25ea33a28  aqua-installer" | sha256sum -c -
RUN chmod +x aqua-installer
RUN ./aqua-installer
