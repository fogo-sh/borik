# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25
ARG DEBIAN_VERSION=trixie
ARG IMAGEMAGICK_VERSION=7.1.2-3

FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS go-deps
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM go-deps AS bot-build
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/borik \
      ./cmd/borik

FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS imagemagick-build
ARG IMAGEMAGICK_VERSION
ENV IMAGEMAGICK_VERSION=${IMAGEMAGICK_VERSION}
COPY scripts/install-imagemagick.sh /usr/local/bin/install-imagemagick
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    install-imagemagick

FROM imagemagick-build AS worker-build
ENV CGO_CFLAGS_ALLOW=-Xpreprocessor \
    PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/borik-worker \
      ./cmd/borik-worker

FROM gcr.io/distroless/static-debian13:nonroot AS bot
WORKDIR /app
COPY --from=bot-build /out/borik /usr/local/bin/borik
ENTRYPOINT ["/usr/local/bin/borik"]

FROM debian:${DEBIAN_VERSION}-slim AS worker
COPY scripts/install-imagemagick.sh /usr/local/bin/install-imagemagick
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt/lists,sharing=locked \
    install-imagemagick --worker-runtime-deps && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd --system --gid 65532 borik && \
    useradd --system --uid 65532 --gid 65532 --home-dir /app --shell /usr/sbin/nologin borik && \
    mkdir -p /app && \
    chown borik:borik /app
COPY --from=imagemagick-build /usr/local /usr/local
RUN ldconfig
COPY --from=worker-build /out/borik-worker /usr/local/bin/borik-worker
WORKDIR /app
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/borik-worker"]
