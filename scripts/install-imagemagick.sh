#!/bin/sh
set -eu

IMAGEMAGICK_VERSION="${IMAGEMAGICK_VERSION:-7.1.2-3}"
PREFIX="${IMAGEMAGICK_PREFIX:-/usr/local}"
WORKDIR="${IMAGEMAGICK_BUILD_DIR:-/tmp/borik-imagemagick-build}"

imagemagick_runtime_packages() {
  echo \
  ghostscript \
  libfontconfig1 \
  libfreetype6 \
  libgif7 \
  libglib2.0-0 \
  libheif1 \
  libjpeg62-turbo \
  liblcms2-2 \
  liblqr-1-0 \
  libpng16-16 \
  libtiff6 \
  libwebp7 \
  libwebpdemux2 \
  libwebpmux3
}

imagemagick_build_packages() {
  echo \
  build-essential \
  ca-certificates \
  libfontconfig1-dev \
  libfreetype6-dev \
  libgif-dev \
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
}

worker_runtime_packages() {
  echo \
  ca-certificates \
  ffmpeg \
  "$(imagemagick_runtime_packages)"
}

install_packages() {
  apt-get update
  apt-get install --yes --no-install-recommends "$@"
}

case "${1:-}" in
  --worker-runtime-deps)
    # shellcheck disable=SC2046
    install_packages $(worker_runtime_packages)
    exit 0
    ;;
  --imagemagick-build-deps)
    # shellcheck disable=SC2046
    install_packages $(imagemagick_runtime_packages) $(imagemagick_build_packages)
    exit 0
    ;;
  "" | --build)
    ;;
  *)
    echo "usage: $0 [--build|--worker-runtime-deps|--imagemagick-build-deps]" >&2
    exit 64
    ;;
esac

# shellcheck disable=SC2046
install_packages $(imagemagick_runtime_packages) $(imagemagick_build_packages)

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
