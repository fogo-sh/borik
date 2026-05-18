FROM golang:1.25-trixie

WORKDIR /deps
ENV IMAGEMAGICK_VERSION=7.1.2-3

COPY scripts/install-imagemagick.sh /usr/local/bin/install-imagemagick
RUN install-imagemagick && \
    apt-get install --yes --no-install-recommends ffmpeg

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build

ENTRYPOINT ["/build/borik", "run"]
