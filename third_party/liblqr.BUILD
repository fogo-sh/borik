load("@rules_foreign_cc//foreign_cc:defs.bzl", "configure_make")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

configure_make(
    name = "liblqr",
    lib_source = ":all",
    configure_options = [
        "--disable-shared",
        "--enable-static",
    ],
    env = {
        # Homebrew provides glib-2.0 (required by liblqr) via pkg-config.
        "PKG_CONFIG_PATH": "/opt/homebrew/lib/pkgconfig:$$EXT_BUILD_DEPS$$/lib/pkgconfig",
        # rules_foreign_cc sets AR=/usr/bin/libtool on macOS; override with real ar.
        "AR": "ar",
    },
    out_static_libs = ["liblqr-1.a"],
    visibility = ["//visibility:public"],
)
