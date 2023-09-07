package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stealthrocket/coroutine/coroc/compiler"
)

const usage = `
coroc is a coroutine compiler for Go.

USAGE:
  coroc [OPTIONS] [PATH]

OPTIONS:
      --output <FILENAME>  Name of the Go file to generate in each package

      --tags   <TAGS...>   Build tags to set on generated files

  -h, --help               Show this help information
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outputFilename := flag.String("output", "", "")
	buildTags := flag.String("tags", "", "")

	flag.Usage = func() { println(usage[1:]) }
	flag.Parse()

	path := flag.Arg(0)
	if path == "" {
		// If the compiler was invoked via go generate, the GOFILE
		// environment variable will be set with the name of the file
		// that contained the go:generate directive, and the current
		// working directory will be set to the directory that
		// contained the file.
		if gofile := os.Getenv("GOFILE"); gofile != "" {
			path = gofile
		} else {
			path = "."
		}
	}

	var options []compiler.CompileOption
	if *outputFilename != "" {
		options = append(options, compiler.WithOutputFilename(*outputFilename))
	}
	if *buildTags != "" {
		options = append(options, compiler.WithBuildTags(*buildTags))
	}

	return compiler.Compile(path, options...)
}
