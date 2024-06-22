ARG GO_VERSION
ARG ALPINE_VERSION

ARG TARGETOS="linux"
ARG TARGETARCH="amd64"

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

ARG DOCKER_META_VERSION=
ARG GITHUB_SHA=
ARG GITHUB_ACTOR=
ARG BUILD_DATE=
ARG GITHUB_BASE_REF=

ENV PROMETHEUS_COMMON_PKG=github.com/prometheus/common

RUN --mount=type=bind,target=/app,source=. <<EOT
    cd /app
    BUILD_DATE=$(date +"%Y%m%d-%T")
    go mod tidy -v
    go build go build -ldflags="-s \
        -X ${PROMETHEUS_COMMON_PKG}/version.Revision=${GITHUB_SHA} \
        -X ${PROMETHEUS_COMMON_PKG}/version.BuildUser=${GITHUB_ACTOR} \
        -X ${PROMETHEUS_COMMON_PKG}/version.BuildDate=${BUILD_DATE} \
        -X ${PROMETHEUS_COMMON_PKG}/version.Branch=${GITHUB_BASE_REF} \
        -X ${PROMETHEUS_COMMON_PKG}/version.Version=${DOCKER_META_VERSION} \
    " \
    -o /prometheus-configs-provider
EOT

FROM quay.io/prometheus/busybox-${TARGETOS}-${TARGETARCH}:latest
COPY --from=builder /prometheus-configs-provider /bin/prometheus-configs-provider
ENTRYPOINT [ "/bin/prometheus-configs-provider" ]
