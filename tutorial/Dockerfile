FROM alpine:3.14.2
RUN apk add curl tree bash
ARG aqua_version
ARG aqua_installer_version
RUN curl -sSfL https://raw.githubusercontent.com/suzuki-shunsuke/aqua-installer/$aqua_installer_version/aqua-installer | bash -s -- -v $aqua_version
