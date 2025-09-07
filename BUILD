load("@gazelle//:def.bzl", "gazelle")
load("@rules_go//go:def.bzl", "go_binary", "go_library")

gazelle(name = "gazelle")

go_library(
    name = "borik_lib",
    srcs = ["main.go"],
    importpath = "github.com/fogo-sh/borik",
    visibility = ["//visibility:private"],
    deps = ["//cmd"],
)

go_binary(
    name = "borik",
    embed = [":borik_lib"],
    visibility = ["//visibility:public"],
)
