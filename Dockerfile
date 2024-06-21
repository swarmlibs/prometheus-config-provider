ARG GO_VERSION
ARG ALPINE_VERSION

ARG TARGETOS="linux"
ARG TARGETARCH="amd64"

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
RUN --mount=type=bind,target=/app,source=. \
    cd /app && go build -o /prometheus-configs-provider

FROM quay.io/prometheus/busybox-${TARGETOS}-${TARGETARCH}:latest
COPY --from=builder /prometheus-configs-provider /bin/prometheus-configs-provider
ENTRYPOINT [ "/bin/prometheus-configs-provider" ]
