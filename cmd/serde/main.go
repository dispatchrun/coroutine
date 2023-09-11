package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"log/slog"
	"os"
	"strings"

	"github.com/stealthrocket/coroutine/serde"
	"golang.org/x/tools/go/packages"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of serde:\n")
	fmt.Fprintf(os.Stderr, "\tserde [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tserde [flags] -type T files...\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

// TODO: this function is duplicated
func enableDebugLogs() {
	removeTime := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		return a
	}

	var programLevel = new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:       programLevel,
		ReplaceAttr: removeTime,
	})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)
}

func main() {
	enableDebugLogs()
	typeName := ""
	flag.StringVar(&typeName, "type", "", "non-optional type name")
	buildTags := ""
	flag.StringVar(&buildTags, "tags", "", "comma-separated list of build tags to apply")
	flag.Usage = usage
	flag.Parse()

	if len(typeName) == 0 {
		fmt.Fprintf(os.Stderr, "missing type name (-type is required)\n")
		flag.Usage()
		os.Exit(2)
	}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	tags := strings.Split(buildTags, ",")

	err := generate(typeName, args, tags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func generate(typeName string, patterns []string, tags []string) error {
	pkgs, err := parse(patterns, tags)
	if err != nil {
		return err
	}

	// Find the package that contains the type declaration requested.
	// This will also be the output package.
	td := serde.FindTypeDef[types.Type]("", typeName, pkgs)
	if td == serde.Notype {
		return fmt.Errorf("could not find type definition")
	}

	output := td.TargetFile()

	g := serde.NewGenerator(tags, pkgs, td.Pkg)
	g.GenRegister(pkgs)

	var buf bytes.Buffer
	n, err := g.WriteTo(&buf)
	if err != nil {
		panic(fmt.Errorf("couldn't write (%d bytes): %w", n, err))
	}

	clean, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println(buf.String())
		return err
	}

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening '%s': %w", output, err)
	}
	defer f.Close()

	_, err = f.Write(clean)

	fmt.Println("[GEN]", output)

	return err
}

func parse(patterns []string, tags []string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}
	return packages.Load(cfg, patterns...)
}
