package serde

import (
	"fmt"
	"go/types"
	"io"
	"math"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

type importsmap struct {
	byName map[string]string
	byPath map[string]string
}

// Add a new import and assign it an imported name, avoiding clashes.
func (m *importsmap) Add(p string) string {
	name, ok := m.byPath[p]
	if ok {
		return name
	}
	if m.byName == nil {
		m.byName = make(map[string]string)
		m.byPath = make(map[string]string)
	}

	original := path.Base(p)
	for i := 0; i <= math.MaxInt; i++ {
		name := original
		if i > 0 {
			name = fmt.Sprintf("%s_%d", original, i)
		}
		_, ok = m.byName[name]
		if ok { // name clash
			continue
		}
		m.byName[name] = p
		m.byPath[p] = name
		return name
	}

	panic("exhausted suffixes")
}

// Ensure path is imported, and its import name matches the one provided. Panics
// otherwise.
func (m *importsmap) Ensure(name, path string) {
	// Since the goal is to panic, it's ok to potentially add an import with
	// the wrong name.
	importname := m.Add(path)
	if importname != name {
		panic(fmt.Errorf("import package '%s' is imported as '%s'; expected '%s'", path, importname, name))
	}
}

type Generator struct {
	// Build tags to decorate the output
	tags []string
	// Type of the serde.Serializable interface.
	serializable *types.Interface
	// Map a package name to its import path.
	imports importsmap

	// Package the output file belongs to.
	main *packages.Package
	// Output.
	s *strings.Builder
}

func NewGenerator(tags []string, pkgs []*packages.Package, target *packages.Package) *Generator {
	// Find our built-in Serializable interface type so that we can check
	// for its implementations.
	serializable := FindTypeDef[*types.Named]("github.com/stealthrocket/coroutine/serde", "Serializable", pkgs)
	if serializable == Notype {
		panic("could not find built-in Serializable interface; make sure coroutine/serde is in pkgs")
	}
	serializableIface := serializable.Obj.Type().(*types.Named).Underlying().(*types.Interface)
	return &Generator{
		tags:         tags,
		serializable: serializableIface,
		main:         target,
	}
}

func (g *Generator) W(f string, args ...any) {
	if g.s == nil {
		g.s = &strings.Builder{}
	}
	fmt.Fprintf(g.s, f, args...)
	g.s.WriteString("\n")
}

func cutLast(s, sep string) (before, after string, found bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return string(s[:i]), string(s[i+1:]), true
}

func public(name string) bool {
	c := name[0] // want to panic if len is 0
	return c >= 'A' && c <= 'Z'
}

func (g *Generator) GenRegister(pkgs []*packages.Package) {
	s := map[string]struct{}{}  // set of types seen
	ps := map[string]struct{}{} // packages already visited

	var findTypes func(p *packages.Package)
	findTypes = func(p *packages.Package) {
		if _, ok := ps[p.ID]; ok {
			return
		}
		ps[p.ID] = struct{}{}

		for _, o := range p.TypesInfo.Defs {
			if o == nil {
				continue
			}
			t := derefType(o.Type())
			name, ok := g.supported(t)
			if !ok {
				continue
			}
			s[name] = struct{}{}
		}

		for _, i := range p.TypesInfo.Instances {
			t := derefType(i.Type)
			name, ok := g.supported(t)
			if !ok {
				continue
			}
			s[name] = struct{}{}
		}

		for _, i := range p.Imports {
			findTypes(i)
		}
	}

	for _, pkg := range pkgs {
		findTypes(pkg)
	}

	g.W(`func init() {`)
	for k := range s {
		g.W(`serde.RegisterType[%s]()`, k)
	}
	g.W(`}`)
}

// return name
func (g *Generator) supported(t types.Type) (name string, ok bool) {
	outpath := g.main.Types.Path()
	if isFunc(t) {
		return "", false
	}
	if isGeneric(t) {
		return "", false
	}
	typename := types.TypeString(t, nil)
	if strings.Contains(typename, "chan ") {
		return "", false
	}
	if strings.Contains(typename, "invalid type") {
		return "", false

	}
	if strings.Contains(typename, "untyped ") {
		return "", false

	}

	path, name, found := cutLast(typename, ".")
	if found {
		if strings.ContainsAny(path, "[]{}*<>") {
			return "", false
		}
		if strings.ContainsAny(name, "[]{}*<>") {
			return "", false
		}
		if strings.Contains(path, "internal/") {
			if !strings.HasPrefix(path, outpath+"/internal/") {
				return "", false
			}
			// TODO: check not target package's internal
			return "", false
		}
		if !public(name) {
			return "", false

		}

		if path != outpath {
			importname := g.imports.Add(path)
			name = importname + "." + name
		}
	} else {
		if strings.ContainsAny(path, "[]{}*<>") {
			return "", false

		}
		name = path
	}
	return name, true
}

func derefType(t types.Type) types.Type {
	switch x := t.(type) {
	case *types.Pointer:
		return derefType(x.Elem())
	default:
		return t
	}
}

func isFunc(t types.Type) bool {
	switch t.(type) {
	case *types.Signature:
		return true
	case *types.Named:
		return isFunc(t.Underlying())
	default:
		return false
	}
}

func isGeneric(t types.Type) bool {
	switch x := t.(type) {
	case *types.Interface:
		return true // ok that's a strech
	case *types.Named:
		if x.Origin() != t {
			return true
		}
		return isGeneric(t.Underlying())
	case *types.TypeParam:
		return true
	default:
		return false
	}
}

func (g *Generator) WriteTo(w io.Writer) (int64, error) {
	var sb strings.Builder
	sb.WriteString("// Code generated by serde. DO NOT EDIT.\n\n")

	if len(g.tags) > 0 {
		tags := strings.Join(g.tags, " ")
		fmt.Fprintf(&sb, "//go:build %s\n", tags)
	}
	fmt.Fprintf(&sb, "package %s\n", g.main.Name)

	n, err := fmt.Fprint(w, sb.String())
	if err != nil {
		return int64(n), err
	}
	for name, path := range g.imports.byName {
		n2, err := fmt.Fprintf(w, "import %s \"%s\"\n", name, path)
		n += n2
		if err != nil {
			return int64(n), err
		}
	}

	n2, err := w.Write([]byte(g.s.String()))
	return int64(n) + int64(n2), err
}

func (g *Generator) TypeNameFor(t types.Type) string {
	return types.TypeString(t, types.RelativeTo(g.main.Types))
}

type Typedef struct {
	Obj types.Object
	Pkg *packages.Package
}

// TargetFile returns the path where a serder function should be generated for
// this type.
func (t Typedef) TargetFile() string {
	pos := t.Pkg.Fset.Position(t.Obj.Pos())
	dir, file := filepath.Split(pos.Filename)

	// Try to preserve build tags in the file name.
	i := strings.LastIndexByte(file, '.')
	if i == -1 {
		panic(fmt.Errorf("files does not end in .go: %s", file))
	}
	noext := file[:i]
	parts := strings.Split(noext, "_")
	parts[0] = "serdegenerated"

	outFile := strings.Join(parts, "_") + ".go"
	return filepath.Join(dir, outFile)
}

var Notype = Typedef{}

func FindTypeDef[T types.Type](inpkg string, name string, pkgs []*packages.Package) Typedef {
	for _, pkg := range pkgs {
		if inpkg != "" && pkg.PkgPath != inpkg {
			continue
		}
		for id, d := range pkg.TypesInfo.Defs {
			if d == nil {
				continue
			}
			_, ok := d.Type().(T)
			if !ok {
				continue
			}
			if id.Name == name {
				return Typedef{Obj: d, Pkg: pkg}
			}
		}
	}
	return Notype
}
