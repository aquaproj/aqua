FROM alpine:3.14.2
RUN apk add curl tree bash
ARG aqua_version
RUN curl -LOq https://github.com/suzuki-shunsuke/aqua/releases/download/$aqua_version/aqua_linux_amd64.tar.gz && \
    tar xvzf aqua_linux_amd64.tar.gz && \
    rm aqua_linux_amd64.tar.gz && \
    rm LICENSE README.md && \
    mv aqua /usr/local/bin/aqua
