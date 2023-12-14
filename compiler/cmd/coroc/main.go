package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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
  -l, --list      List all files that would be compiled
  -v, --version   Show the compiler version
`

var (
	showVersion   bool
	onlyListFiles bool
)

func boolFlag(ptr *bool, short, long string) {
	flag.BoolVar(ptr, short, false, "")
	flag.BoolVar(ptr, long, false, "")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = func() { println(usage[1:]) }

	boolFlag(&showVersion, "v", "version")
	boolFlag(&onlyListFiles, "l", "list")
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

	if onlyListFiles {
		log.SetOutput(io.Discard)
	}

	return compiler.Compile(path,
		compiler.OnlyListFiles(onlyListFiles),
	)
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
