load("@rules_foreign_cc//foreign_cc:defs.bzl", "cmake")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

cmake(
    name = "libjpeg_turbo",
    lib_source = ":all",
    out_static_libs = ["libjpeg.a"],
    cache_entries = {
        "BUILD_SHARED_LIBS": "OFF",
        "CMAKE_POSITION_INDEPENDENT_CODE": "ON",
        "WITH_TURBOJPEG": "FALSE",
    },
    visibility = ["//visibility:public"],
)
