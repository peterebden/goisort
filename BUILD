go_get(
    name = "go-flags",
    get = "github.com/jessevdk/go-flags",
    revision = "v1.4.0",
)

go_binary(
    name = "goisort",
    srcs = ["main.go"],
    deps = [
        ":go-flags",
        "//isort",
    ],
)
