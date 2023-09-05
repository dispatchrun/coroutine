package main

import (
	"flag"
	"fmt"
	"go/types"
	"io"
	"os"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of serde:\n")
	fmt.Fprintf(os.Stderr, "\tserde [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tserde [flags] -type T files...\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	typeName := ""
	flag.StringVar(&typeName, "type", "", "non-optional type name")
	output := ""
	flag.StringVar(&output, "output", "", "output file name; defaults to <type_serde.go")
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

	err := generate(typeName, args, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
}

func generate(typeName string, patterns []string, output string) error {
	pkgs, err := parse(patterns)
	if err != nil {
		return err
	}

	main := pkgs[0]

	set := findAllTypes(pkgs)

	fprintTypeSet(os.Stdout, main, set)

	fmt.Println("types:", set.Len())

	return nil
}

// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func parse(patterns []string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
	}
	return packages.Load(cfg, patterns...)
}

type pkgSet map[string]struct{}

func (s pkgSet) Has(p *packages.Package) bool {
	_, ok := s[p.ID]
	return ok
}

func (s pkgSet) Set(p *packages.Package) {
	s[p.ID] = struct{}{}
}

type typeSet struct {
	m *typeutil.Map
}

func (s *typeSet) Set(t types.Type) {
	if s.m == nil {
		s.m = &typeutil.Map{}
	}
	s.m.Set(t, struct{}{})
}

func (s *typeSet) Iterate(f func(t types.Type)) {
	s.m.Iterate(func(key types.Type, _ interface{}) {
		f(key)
	})
}

func (s *typeSet) Len() int {
	return s.m.Len()
}

func fprintTypeSet(w io.Writer, main *packages.Package, s typeSet) {
	s.Iterate(func(t types.Type) {
		extra := typeStr(main.Types, t)
		fmt.Fprintf(w, "- %-70s | %s\n", t.String(), extra)
	})
}

func typeStr(main *types.Package, t types.Type) string {
	return types.TypeString(t, types.RelativeTo(main))
}

func basicKindString(x *types.Basic) string {
	names := [...]string{
		types.Invalid:       "Invalid",
		types.Bool:          "Bool",
		types.Int:           "Int",
		types.Int8:          "Int8",
		types.Int16:         "Int16",
		types.Int32:         "Int32",
		types.Int64:         "Int64",
		types.Uint:          "Uint",
		types.Uint8:         "Uint8",
		types.Uint16:        "Uint16",
		types.Uint32:        "Uint32",
		types.Uint64:        "Uint64",
		types.Uintptr:       "Uintptr",
		types.Float32:       "Float32",
		types.Float64:       "Float64",
		types.Complex64:     "Complex64",
		types.Complex128:    "Complex128",
		types.String:        "String",
		types.UnsafePointer: "UnsafePointer",

		types.UntypedBool:    "UntypedBool",
		types.UntypedInt:     "UntypedInt",
		types.UntypedRune:    "UntypedRune",
		types.UntypedFloat:   "UntypedFloat",
		types.UntypedComplex: "UntypedComplex",
		types.UntypedString:  "UntypedString",
		types.UntypedNil:     "UntypedNil",
	}

	return names[x.Kind()]
}

func findAllTypes(pkgs []*packages.Package) typeSet {
	var s typeSet
	var p pkgSet = make(pkgSet)
	for _, pkg := range pkgs {
		s = findTypes(p, s, pkg)
	}
	return s
}

func findTypes(pkgs pkgSet, s typeSet, p *packages.Package) typeSet {
	if pkgs.Has(p) {
		return s
	}
	pkgs.Set(p)
	for _, o := range p.TypesInfo.Defs {
		if o == nil {
			continue
		}
		t := o.Type()
		s.Set(t)
	}

	for _, i := range p.Imports {
		s = findTypes(pkgs, s, i)
	}

	return s
}
