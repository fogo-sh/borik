FROM golang:1.18-buster

WORKDIR /deps
ENV IMAGEMAGICK_VERSION=7.1.0-57

RUN apt-get update && \
    apt-get -q -y install libjpeg-dev libpng-dev libtiff-dev libgif-dev libx11-dev liblqr-1-0-dev --no-install-recommends && \
    wget https://github.com/ImageMagick/ImageMagick/archive/${IMAGEMAGICK_VERSION}.tar.gz && \
	tar xvzf ${IMAGEMAGICK_VERSION}.tar.gz && \
	cd ImageMagick* && \
	./configure \
	    --without-magick-plus-plus \
	    --without-perl \
	    --disable-openmp \
	    --with-gvc=no \
	    --disable-docs \
        --with-lqr=yes && \
	make -j$(nproc) && make install && \
	ldconfig /usr/local/lib

WORKDIR /build
COPY . .
RUN go mod download &&\
    go mod verify &&\
    go build

ENTRYPOINT ["/build/borik"]
