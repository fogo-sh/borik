FROM golang:1.18-bullseye

WORKDIR /deps
ENV IMAGEMAGICK_VERSION=7.1.0-57

RUN apt-get update && \
    apt-get -q -y install \
      # File formats
      libjpeg-dev \
      libpng-dev \
      libtiff-dev \
      libgif-dev \
      libwebp-dev \
      libheif-dev \
      ffmpeg \
      # Operations
      liblqr-1-0-dev \
      libglib2.0 \
      # Other
      libx11-dev \
      --no-install-recommends && \
    wget https://github.com/ImageMagick/ImageMagick/archive/${IMAGEMAGICK_VERSION}.tar.gz && \
	tar xvzf ${IMAGEMAGICK_VERSION}.tar.gz && \
	cd ImageMagick* && \
	./configure && \
	make -j$(nproc) && make install && \
	ldconfig /usr/local/lib

WORKDIR /build
COPY . .
RUN go mod download &&\
    go mod verify &&\
    go build

ENTRYPOINT ["/build/borik"]
