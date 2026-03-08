load("@rules_foreign_cc//foreign_cc:defs.bzl", "cmake")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

cmake(
    name = "libwebp",
    cache_entries = {
        "BUILD_SHARED_LIBS": "OFF",
        "CMAKE_POSITION_INDEPENDENT_CODE": "ON",
        "WEBP_BUILD_ANIM_UTILS": "OFF",
        "WEBP_BUILD_CWEBP": "OFF",
        "WEBP_BUILD_DWEBP": "OFF",
        "WEBP_BUILD_EXTRAS": "OFF",
        "WEBP_BUILD_GIF2WEBP": "OFF",
        "WEBP_BUILD_IMG2WEBP": "OFF",
        "WEBP_BUILD_VWEBP": "OFF",
        "WEBP_BUILD_WEBPINFO": "OFF",
        "WEBP_BUILD_WEBPMUX": "ON",
    },
    lib_source = ":all",
    out_static_libs = [
        "libwebp.a",
        "libsharpyuv.a",
        "libwebpmux.a",
        "libwebpdemux.a",
    ],
    visibility = ["//visibility:public"],
)
