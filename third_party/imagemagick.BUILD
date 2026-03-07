load("@rules_foreign_cc//foreign_cc:defs.bzl", "configure_make")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

configure_make(
    name = "imagemagick",
    lib_source = ":all",
    deps = [
        "@zlib_src//:zlib",
        "@libjpeg_turbo_src//:libjpeg_turbo",
        "@libpng_src//:libpng",
        "@giflib_src//:giflib",
        "@libwebp_src//:libwebp",
        "@freetype_src//:freetype",
        "@liblqr_src//:liblqr",
    ],
    configure_options = [
        "--enable-static",
        "--disable-shared",
        "--enable-hdri",
        "--with-quantum-depth=16",
        "--without-x",
        "--without-openjp2",
        "--without-gvc",
        "--without-lzma",
        "--without-lcms",
        "--without-jxl",
        "--without-fontconfig",
        "--without-xml",
        "--without-tiff",
        "--without-heic",
        "--without-openexr",
        "--without-raw",
        "--without-perl",
        "--without-magick-plus-plus",
        "--disable-openmp",
        "--disable-docs",
        "--with-jpeg=yes",
        "--with-png=yes",
        "--with-gif=yes",
        "--with-webp=yes",
        "--with-freetype=yes",
        "--with-lqr=yes",
    ],
    env = {
        # rules_foreign_cc sets AR=/usr/bin/libtool on macOS; override with real ar.
        "AR": "ar",
        # Point configure to the dep trees built by rules_foreign_cc.
        "CPPFLAGS": "-I$$EXT_BUILD_DEPS$$/include -I$$EXT_BUILD_DEPS$$/include/freetype2",
        "LDFLAGS": "-L$$EXT_BUILD_DEPS$$/lib -lsharpyuv",
        # Homebrew provides glib (for liblqr); remaining vars are autoconf-style
        # overrides so configure picks up our built deps without needing working .pc files.
        "PKG_CONFIG_PATH": "/opt/homebrew/lib/pkgconfig:$$EXT_BUILD_DEPS$$/lib/pkgconfig",
        "JPEG_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include",
        "JPEG_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -ljpeg",
        "PNG_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include",
        "PNG_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -lpng",
        "GIF_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include",
        "GIF_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -lgif",
        "WEBP_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include",
        "WEBP_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -lwebp -lsharpyuv -lwebpmux -lwebpdemux",
        "FREETYPE_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include/freetype2",
        "FREETYPE_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -lfreetype",
        "LQR_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include/lqr-1",
        "LQR_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -llqr-1",
    },
    out_include_dir = "include/ImageMagick-7",
    out_static_libs = [
        "libMagickWand-7.Q16HDRI.a",
        "libMagickCore-7.Q16HDRI.a",
    ],
    visibility = ["//visibility:public"],
)
