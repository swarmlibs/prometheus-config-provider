ARG GO_VERSION
ARG ALPINE_VERSION

# buildkit
ARG TARGETOS="linux"
ARG TARGETARCH="amd64"
ARG BUILDPLATFORM="linux/amd64"

# docker-metadata-action
ARG DOCKER_META_VERSION=

# github-metadata-action
ARG GITHUB_SHA=
ARG GITHUB_ACTOR=
ARG BUILD_DATE=
ARG GITHUB_BASE_REF=

FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
ARG DOCKER_META_VERSION
ARG GITHUB_SHA
ARG GITHUB_ACTOR
ARG BUILD_DATE
ARG GITHUB_BASE_REF
RUN --mount=type=bind,target=/app,source=. \
    --mount=type=cache,target=/go/pkg/mod \
<<EOT
    set -ex
    cd /app
    export PROMETHEUS_COMMON_PKG=github.com/prometheus/common
    export BUILD_DATE=$(date +"%Y%m%d-%T")
    export CGO_ENABLED=0
    export GOOS=linux
    for GOARCH in amd64 arm64; do
        export GOARCH
        go build -o /prometheus-configs-provider-$GOOS-$GOARCH \
        -ldflags="-s \
            -X ${PROMETHEUS_COMMON_PKG}/version.Revision=${GITHUB_SHA} \
            -X ${PROMETHEUS_COMMON_PKG}/version.BuildUser=${GITHUB_ACTOR} \
            -X ${PROMETHEUS_COMMON_PKG}/version.BuildDate=${BUILD_DATE} \
            -X ${PROMETHEUS_COMMON_PKG}/version.Branch=${GITHUB_BASE_REF} \
            -X ${PROMETHEUS_COMMON_PKG}/version.Version=${DOCKER_META_VERSION} \
        "
    done
EOT

FROM quay.io/prometheus/busybox-${TARGETOS}-${TARGETARCH}:latest
ARG TARGETOS
ARG TARGETARCH
COPY --from=builder /prometheus-configs-provider-$TARGETOS-$TARGETARCH /bin/prometheus-configs-provider
ENTRYPOINT [ "/bin/prometheus-configs-provider" ]
