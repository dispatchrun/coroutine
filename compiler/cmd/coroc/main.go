package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/stealthrocket/coroutine/compiler"
)

const usage = `
coroc is a coroutine compiler for Go.

USAGE:
  coroc [OPTIONS] [PATH]

OPTIONS:
  -h, --help      Show this help information
  -v, --version   Show the compiler version
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = func() { println(usage[1:]) }

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "")
	flag.BoolVar(&showVersion, "version", false, "")

	flag.Parse()

	if showVersion {
		fmt.Println(version())
		return nil
	}

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

	return compiler.Compile(path)
}

func version() (version string) {
	version = "devel"
	if info, ok := debug.ReadBuildInfo(); ok {
		switch info.Main.Version {
		case "":
		case "(devel)":
		default:
			version = info.Main.Version
		}
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				version += " " + setting.Value
			}
		}
	}
	return
}
