package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"runtime/pprof"

	"github.com/dispatchrun/coroutine/compiler"
)

const usage = `
coroc is a coroutine compiler for Go.

USAGE:
  coroc [OPTIONS] [PATH]

OPTIONS:
  -h, --help         Show this help information
  -l, --list         List all files that would be compiled
  -v, --version      Show the compiler version

ADVANCED OPTIONS:
  -callgraph <TYPE>  Set the callgraph construction algorithm
                     (static, cha, rta, vta). Default is vta.

  -colors            Print debug information about function colors

  -cpuprofile        Write CPU profile to file
  -memprofile        Write memory profile to file
`

var (
	showVersion   bool
	onlyListFiles bool
	debugColors   bool
	callgraphType string
	cpuProfile    string
	memProfile    string
)

func boolFlag(ptr *bool, names ...string) {
	for _, name := range names {
		flag.BoolVar(ptr, name, false, "")
	}
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
	boolFlag(&debugColors, "colors")
	flag.StringVar(&callgraphType, "callgraph", "", "")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "")
	flag.StringVar(&memProfile, "memprofile", "", "")
	flag.Parse()

	if showVersion {
		fmt.Println(version())
		return nil
	}

	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			return err
		}
		defer f.Close()
		defer func() {
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Printf("warning: failed to write heap profile: %v", err)
			}
		}()
	}

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
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
		compiler.CallgraphType(callgraphType),
		compiler.OnlyListFiles(onlyListFiles),
		compiler.DebugColors(debugColors),
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
