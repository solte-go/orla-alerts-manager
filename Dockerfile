# syntax=docker/dockerfile:1.4

FROM golang:1.19-alpine3.16 AS builder
RUN --mount=type=cache,target=/root/.cache \
    apk add --no-cache --update git

WORKDIR /build
COPY --link go.mod go.sum ./
COPY --link cmd cmd
COPY --link src src
COPY --link vendor vendor

ARG CGO_ENABLED=0
ARG GOOS=linux

ARG CMDPATH=soltelab/rabbitmq/cmd

RUN  --mount=type=cache,target=/root/.cache \
    go build -tags musl -o ./bin/proxy -ldflags "-X main.version=0.1.0" $CMDPATH/proxy; 


FROM alpine:3.16 as base-container
COPY --link --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --link --from=builder /build/cmd/production.toml /app/cmd/production.toml

# PROXY builder
FROM base-container as proxy
WORKDIR /app
COPY --link --from=builder /build/bin/proxy /app/cmd/proxy

EXPOSE 5005

WORKDIR /app/cmd
ENTRYPOINT ["./proxy", "-env", "prod"]
