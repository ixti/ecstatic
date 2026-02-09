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
      -ldflags="-s -w -X github.com/ixti/ecs-task-helper/cmd.version=${VERSION}" \
      -o /ecs-task-helper \
      .

FROM scratch
COPY --from=builder /ecs-task-helper /ecs-task-helper

ENTRYPOINT ["/ecs-task-helper"]
CMD ["help"]
