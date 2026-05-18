#!/usr/bin/env bash
set -euo pipefail

IMAGEMAGICK_VERSION="${IMAGEMAGICK_VERSION:-7.1.2-3}"
WORKDIR="${IMAGEMAGICK_BUILD_DIR:-/tmp/borik-imagemagick-build}"

apt-get update
apt-get install --yes --no-install-recommends \
  build-essential \
  ghostscript \
  libfontconfig1-dev \
  libfreetype6-dev \
  libgif-dev \
  libglib2.0-0 \
  libglib2.0-dev \
  libheif-dev \
  libjpeg-dev \
  liblcms2-dev \
  liblqr-1-0-dev \
  libpng-dev \
  libtiff-dev \
  libwebp-dev \
  pkg-config \
  wget

rm -rf "${WORKDIR}"
mkdir -p "${WORKDIR}"
cd "${WORKDIR}"

wget "https://github.com/ImageMagick/ImageMagick/archive/${IMAGEMAGICK_VERSION}.tar.gz"
tar xzf "${IMAGEMAGICK_VERSION}.tar.gz"
cd "ImageMagick-${IMAGEMAGICK_VERSION}"

./configure
make -j"$(nproc)"
make install
ldconfig /usr/local/lib

rm -rf "${WORKDIR}"
