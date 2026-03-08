load("@rules_foreign_cc//foreign_cc:defs.bzl", "configure_make")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

configure_make(
    name = "liblqr",
    configure_options = [
        "--disable-shared",
        "--enable-static",
    ],
    env = {
        "GLIB_CFLAGS": "-I$$EXT_BUILD_DEPS$$/include",
        "GLIB_LIBS": "-L$$EXT_BUILD_DEPS$$/lib -lglib-2.0",
        # rules_foreign_cc sets AR=/usr/bin/libtool on macOS; override with real ar.
        "AR": "ar",
    },
    lib_source = ":all",
    out_static_libs = ["liblqr-1.a"],
    visibility = ["//visibility:public"],
    deps = [
        "@glib//glib",
    ],
)
