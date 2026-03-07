load("@rules_foreign_cc//foreign_cc:defs.bzl", "cmake")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

cmake(
    name = "freetype",
    lib_source = ":all",
    deps = [
        "@zlib_src//:zlib",
        "@libpng_src//:libpng",
    ],
    out_static_libs = ["libfreetype.a"],
    cache_entries = {
        "BUILD_SHARED_LIBS": "OFF",
        "CMAKE_POSITION_INDEPENDENT_CODE": "ON",
        "FT_DISABLE_BZIP2": "ON",
        "FT_DISABLE_HARFBUZZ": "ON",
        "FT_DISABLE_BROTLI": "ON",
        "FT_REQUIRE_ZLIB": "ON",
        "FT_REQUIRE_PNG": "ON",
    },
    visibility = ["//visibility:public"],
)
