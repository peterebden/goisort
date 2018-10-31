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

go_get(
    name = "testify",
    get = "github.com/stretchr/testify",
    install = [
        "assert",
        "require",
        "vendor/github.com/davecgh/go-spew/spew",
        "vendor/github.com/pmezard/go-difflib/difflib",
    ],
    test_only = True,
    visibility = ["PUBLIC"],
    #revision = "f390dcf405f7b83c997eac1b06768bb9f44dec18",
    #deps = [":spew"],
)
