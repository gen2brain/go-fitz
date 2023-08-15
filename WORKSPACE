workspace(name = "org_byted_code_x".replace("/", "_"))

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazelize",
    type = "tar.gz",
    urls = [
        "https://{}/repository/scm/api/v1/download_latest/?name=bazel/bazelize/rules".format(
            x,
        )
        for x in [
            "luban-source.byted.org",
            "luban-source.tiktokd.org",
        ]
    ],
)

load("@bazelize//rules/common:common_deps.bzl", "common_deps")

common_deps()

# ----------------------------------------------------
# Gazelle
# ----------------------------------------------------

load("//:go_deps.bzl", "go_dependencies")

# gazelle:repository_macro go_deps.bzl%go_dependencies
go_dependencies()

# This must be invoked after our explicit dependencies
# See https://github.com/bazelbuild/bazel-gazelle/issues/1115.

load("@bazelize//rules/common:workspace_init.bzl", "workspace_init")

workspace_init(go_version = "1.20")
