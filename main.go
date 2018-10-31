// Package main implements goisort, a small and opinionated Go import sorter.
package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/peterebden/goisort/isort"
)

var opts struct {
	LocalPackage string `long:"local_package" short:"l" description:"Import path of the local package (e.g. github.com/peterebden/goisort"`
	Write        bool   `long:"write" short:"w" description:"Rewrite the files in-place"`
	Args         struct {
		Files []flags.Filename `positional-arg-name:"files" required:"true" description:"Files to sort imports in"`
	} `positional-args:"true"`
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	for _, filename := range opts.Args.Files {
		changes, err := isort.Reformat(string(filename), opts.LocalPackage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse %s: %s", filename, err)
			os.Exit(1)
		}
		if opts.Write {
			if err := isort.Rewrite(string(filename), changes); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to rewrite %s: %s", filename, err)
				os.Exit(1)
			}
		}
	}
}
