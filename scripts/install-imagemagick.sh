#!/usr/bin/env bash
set -euo pipefail

IMAGEMAGICK_VERSION="${IMAGEMAGICK_VERSION:-7.1.2-3}"
PREFIX="${IMAGEMAGICK_PREFIX:-/usr/local}"
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

if [ -x "${PREFIX}/bin/magick" ] && "${PREFIX}/bin/magick" -version | grep -q "ImageMagick ${IMAGEMAGICK_VERSION}"; then
  ldconfig "${PREFIX}/lib"
  exit 0
fi

rm -rf "${WORKDIR}"
mkdir -p "${WORKDIR}"
cd "${WORKDIR}"

wget "https://github.com/ImageMagick/ImageMagick/archive/${IMAGEMAGICK_VERSION}.tar.gz"
tar xzf "${IMAGEMAGICK_VERSION}.tar.gz"
cd "ImageMagick-${IMAGEMAGICK_VERSION}"

./configure --prefix="${PREFIX}"
make -j"$(nproc)"
make install
ldconfig "${PREFIX}/lib"

rm -rf "${WORKDIR}"
