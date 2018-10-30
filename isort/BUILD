go_library(
    name = "isort",
    srcs = [
        "isort.go",
        ":packages",
    ],
    visibility = ["PUBLIC"],
)

genrule(
    name = "packages",
    outs = ["packages.go"],
    cmd = [
        "echo 'package isort' > $OUT",
        "echo 'var stdlib = []string{' >> $OUT",
        "eval `go env`",
        "cd $GOROOT",
        "find src -type d | grep -v vendor/ | grep -v internal/ | grep -v cmd/ | sort | cut -c 5- | sed -e 's/^/\"/g' -e 's/$/\",/g' >> $OUT",
        "echo '}' >> $OUT",
    ],
)
