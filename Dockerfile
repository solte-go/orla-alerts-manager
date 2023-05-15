# syntax=docker/dockerfile:1.4
# BUILDER
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

RUN  --mount=type=cache,target=/root/.cache \
    go build -tags musl -o cmd/proxy -ldflags "-X main.version=$VERSION" ../proxy;


# BASE-CONTAINER builder
FROM alpine:3.16 as base-container
COPY --link --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --link --from=builder /build/cmd/production.toml /opt/worker/production.toml

# PROXY builder
FROM base-container as proxy
WORKDIR /app
COPY --link --from=builder /build/cmd/proxy /opt/worker/

EXPOSE 8080

WORKDIR /opt/worker
ENTRYPOINT ["./proxy", "-env", "prod"]
