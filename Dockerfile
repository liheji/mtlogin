FROM golang:1.25-alpine AS builder
RUN go env -w GO111MODULE=auto   && go env -w CGO_ENABLED=0

WORKDIR /build

COPY ./ .

RUN set -ex &&  \
    cd /build &&  \
    go build -ldflags "-s -w -extldflags '-static'" -o mtlogin .

FROM alpine:latest

COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh &&  \
    apk add --no-cache --update coreutils shadow su-exec tzdata &&  \
    rm -rf /var/cache/apk/* &&  \
    mkdir -p /app &&  \
    mkdir -p /app/data &&  \
    mkdir -p /app/conf &&  \
    mkdir -p /app/logs &&  \
    useradd -d /app/conf -s /bin/sh abc &&  \
    chown -R abc /app

ENV TZ="Asia/Shanghai"
ENV UID=99
ENV GID=100
ENV UMASK=002

COPY --from=builder /build/mtlogin /app/
COPY --chown=abc:abc conf/*.toml.tpl /app/conf/

WORKDIR /app

VOLUME [ "/app/data", "/app/conf", "/app/logs" ]

ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD [ "/app/mtlogin" ]
