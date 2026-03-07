load("@rules_foreign_cc//foreign_cc:defs.bzl", "make")

filegroup(
    name = "all",
    srcs = glob(["**"]),
    visibility = ["//visibility:public"],
)

# giflib uses a hand-written Makefile (no autotools/cmake).
make(
    name = "giflib",
    lib_source = ":all",
    out_static_libs = ["libgif.a"],
    targets = ["libgif.a", "install"],
    # rules_foreign_cc sets AR=/usr/bin/libtool on macOS; override with real ar.
    args = ["AR=ar"],
    visibility = ["//visibility:public"],
)
