load("@rules_foreign_cc//foreign_cc:defs.bzl", "cmake")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

cmake(
    name = "libpng",
    lib_source = ":all",
    deps = ["@zlib_src//:zlib"],
    out_static_libs = ["libpng.a"],
    cache_entries = {
        "BUILD_SHARED_LIBS": "OFF",
        "CMAKE_POSITION_INDEPENDENT_CODE": "ON",
    },
    visibility = ["//visibility:public"],
)
