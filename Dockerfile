# Build
FROM golang:1.16-alpine AS build

WORKDIR /usr/app
ADD . /usr/app

RUN apk add --no-cache --update make \
    && rm -f /var/cache/apk/*

RUN go build -o auto-semver-tag

# Runtime
FROM alpine:latest

WORKDIR /usr/app

COPY entrypoint.sh /usr/app/entrypoint.sh
COPY --from=build /usr/app/auto-semver-tag /usr/app/auto-semver-tag

ENTRYPOINT ["/entrypoint.sh"]