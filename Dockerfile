FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN set -eux; \
  go mod download

COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN set -eux; \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
      -ldflags="-s -w -X github.com/ixti/ecstatic/cmd.version=${VERSION}" \
      -o /ecstatic \
      .

FROM scratch

LABEL org.opencontainers.image.title="ecstatic"
LABEL org.opencontainers.image.description="Joyful ECS Task Utilities"
LABEL org.opencontainers.image.url="https://github.com/ixti/ecstatic"
LABEL org.opencontainers.image.source="https://github.com/ixti/ecstatic"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="ixti"

COPY --from=builder /ecstatic /ecstatic

ENTRYPOINT ["/ecstatic"]
CMD ["help"]
