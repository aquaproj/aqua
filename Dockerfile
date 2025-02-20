FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c
COPY dist/aqua-docker /usr/local/bin/aqua
RUN apk add curl bash sudo git vim
RUN adduser -u 1000 -G wheel -D foo
RUN sed -i 's|# %wheel ALL=(ALL:ALL) NOPASSWD|%wheel ALL=(ALL:ALL) NOPASSWD|' /etc/sudoers
USER foo
RUN mkdir /home/foo/workspace
WORKDIR /home/foo/workspace
ENV PATH=/home/foo/.local/share/aquaproj-aqua/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
