package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/stealthrocket/coroutine/serde"
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
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func generate(typeName string, patterns []string, output string) error {
	// Add the serde support library to the search to bring the built-ins
	// into the type system. At the moment it's only used for the
	// Serializable interface, but eventually it should be used to reference
	// helpers and basic types serialization functions by their ast.Ident
	// directly.
	patterns = append(patterns, "github.com/stealthrocket/coroutine/serde")

	pkgs, err := parse(patterns)
	if err != nil {
		return err
	}

	// Find our built-in Serializable interface type so that we can check
	// for its implementations.
	serializable := findTypeDef("Serializable", pkgs)
	if serializable == notype {
		return fmt.Errorf("could not find built-in Serializable interface")
	}
	serializableIface := serializable.obj.Type().(*types.Named).Underlying().(*types.Interface)

	// Find the package that contains the type declaration requested.
	// This will also be the output package.
	td := findTypeDef(typeName, pkgs)
	if td == notype {
		return fmt.Errorf("could not find type definition")
	}

	output = td.TargetFile()

	g := generator{
		serializable: serializableIface,
		output:       td.TargetFile(),
		main:         td.pkg,
	}

	g.Typedef(td)
	g.AllTypes(pkgs)

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
	//	fmt.Println(string(clean))

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening '%s': %w", output, err)
	}
	defer f.Close()

	_, err = f.Write(clean)

	fmt.Println("[GEN]", output)

	return err
}

type location struct {
	pkg      string
	name     string
	terminal bool // if true, doesn't need the serializer/deserializer argument
}

type locations struct {
	serializer   location
	deserializer location
}

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

type generator struct {
	// Type of the serde.Serializable interface.
	serializable *types.Interface
	// Map[types.Type] -> locations to track the types that already have
	// their serialization functions emitted.
	known typeutil.Map
	// Map a package name to its import path.
	imports importsmap

	// Path where the code should be written.
	output string
	// Package the output file belongs to.
	main *packages.Package
	// Output.
	s *strings.Builder
}

func (g *generator) W(f string, args ...any) {
	if g.s == nil {
		g.s = &strings.Builder{}
	}
	fmt.Fprintf(g.s, f, args...)
	g.s.WriteString("\n")
}

// Generate the code for a given typedef
func (g *generator) Typedef(t typedef) {
	typeName := g.TypeNameFor(t.obj.Type())
	g.Type(t.obj.Type(), typeName)
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

func (g *generator) AllTypes(pkgs []*packages.Package) {
	// Generate a reflect.Type <> ID mapping.
	//	s := typeutil.Map{}
	s := map[string]struct{}{}
	ps := map[string]struct{}{}

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
			//	fullname := types.TypeString(t, nil)
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

	g.W(`var _typemap *serde.TypeMap = serde.NewTypeMap(`)
	for k := range s {
		// TODO: find a way to avoid the allocations
		g.W(`reflect.TypeOf(new(%s)).Elem(),`, k)
	}
	g.W(`)`)
}

// return name
func (g *generator) supported(t types.Type) (name string, ok bool) {
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
	//			typename := g.TypeNameFor(t)
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
		//fmt.Println("PATH:", path, "NAME:", name)

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

func (g *generator) WriteTo(w io.Writer) (int64, error) {
	n, err := fmt.Fprintf(w, "// Code generated by coroc. DO NOT EDIT.\n\npackage %s\n", g.main.Name)
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

func (g *generator) Type(t types.Type, name string) locations {
	// Hit the cache first.
	if loc, ok := g.get(t); ok {
		return loc
	}

	if types.AssignableTo(t, g.serializable) {
		return g.Serializable(t, name)
	}

	if types.AssignableTo(types.NewPointer(t), g.serializable) {
		return g.SerializableToPtr(t, name)
	}

	switch x := t.(type) {
	case *types.Basic:
		return g.Basic(x, name)
	case *types.Struct:
		return g.Struct(x, name)
	case *types.Named:
		return g.Named(x, name)
	case *types.Slice:
		return g.Slice(x, name)
	case *types.Pointer:
		return g.Pointer(x, name)
	case *types.Map:
		return g.Map(x, name)
	default:
		panic(fmt.Errorf("type generator not implemented: %s (%T)", t, t))
	}
}

func (g *generator) Map(t *types.Map, name string) locations {
	loc := g.newGenLocation(t, name)

	kt := t.Key()
	kname := g.TypeNameFor(kt)
	kloc := g.Type(kt, kname)

	vt := t.Elem()
	vname := g.TypeNameFor(vt)
	vloc := g.Type(vt, vname)

	g.W(`func %s(s *serde.Serializer, z %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	g.W(`b = serde.SerializeMapSize(z, b)`)
	g.W(`for k, v := range z {`)
	g.W(`{`)
	g.W(`x := k`)
	g.serializeCallForLoc(kloc)
	g.W(`}`)

	g.W(`{`)
	g.W(`x := v`)
	g.serializeCallForLoc(vloc)
	g.W(`}`)

	g.W(`}`)

	g.W(`return b`)
	g.W(`}`)
	g.W(``)

	g.W(`func %s(d *serde.Deserializer, b []byte) (%s, []byte) {`, loc.deserializer.name, name)
	g.W(`d = serde.EnsureDeserializer(d)`)
	g.W(`n, b := serde.DeserializeMapSize(b)`)
	g.W(`var z %s`, name)
	g.W(`if n < 0 { return z, b }`)
	g.W(`z = make(%s, n)`, name)
	g.W(`var k %s`, kname)
	g.W(`var v %s`, vname)
	g.W(`for i := 0; i < n; i++ {`)
	g.W(`{`)
	g.W(`var x %s`, kname)
	g.deserializeCallForLoc(kloc)
	g.W(`k = x`)
	g.W(`}`)
	g.W(`{`)
	g.W(`var x %s`, vname)
	g.deserializeCallForLoc(vloc)
	g.W(`v = x`)
	g.W(`}`)
	g.W(`z[k] = v`)
	g.W(`}`)
	g.W(`return z, b`)
	g.W(`}`)

	return loc
}

func (g *generator) Pointer(t *types.Pointer, name string) locations {
	unsafeName := g.imports.Add("unsafe")
	loc := g.newGenLocation(t, name)

	pt := t.Elem()
	ploc := g.Type(pt, name)
	ptype := g.TypeNameFor(pt)

	g.W(`func %s(s *serde.Serializer, z %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	g.W(`ok, b := s.WritePtr(%s.Pointer(z), b)`, unsafeName)
	g.W(`if !ok {`)
	g.W(`x := *z`)
	g.serializeCallForLoc(ploc)
	g.W(`}`)
	g.W(`return b`)
	g.W(`}`)
	g.W(``)

	g.W(`func %s(d *serde.Deserializer, b []byte) (%s, []byte) {`, loc.deserializer.name, name)
	g.W(`d = serde.EnsureDeserializer(d)`)
	g.W(`p, i, b := d.ReadPtr(b)`)
	g.W(`if p != nil || i == 0 {`)
	g.W(`return (%s)(p), b`, name)
	g.W(`}`)
	// Little dance to create the placeholder pointer for circular
	// references. Would be better if deserialization functions took a
	// pointer argument, which is a TODO.
	g.W(`var x %s`, ptype)
	g.W(`var xx %s`, ptype)
	g.W(`pxx := &xx`)
	g.W(`d.Store(i, %s.Pointer(pxx))`, unsafeName)
	g.deserializeCallForLoc(ploc)
	g.W(`*pxx=x`)
	g.W(`return pxx, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *generator) Serializable(t types.Type, name string) locations {
	return g.builtin(t, "SerializeSerializable", "DeserializeSerializable")
}

func (g *generator) SerializableToPtr(t types.Type, name string) locations {
	// t is not Serializable, but *t is.
	loc := g.newGenLocation(t, name)

	// location for the pointer type
	ploc := g.Type(types.NewPointer(t), name)

	// generate wrappers to use the pointer type
	g.W(`func %s(s *serde.Serializer, z %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	g.W(`x := &z`)
	g.serializeCallForLoc(ploc)
	g.W(`return b`)
	g.W(`}`)
	g.W(``)

	g.W(`func %s(d *serde.Deserializer, b []byte) (%s, []byte) {`, loc.deserializer.name, name)
	g.W(`d = serde.EnsureDeserializer(d)`)
	g.W(`var z %s`, name)
	g.W(`x := &z`)
	// This is a special call because it takes a pointer as target instead
	// of returning the value.
	// TODO: make all signatures like that.
	g.W(`b = serde.DeserializeSerializable(x, b)`)
	g.W(`return z, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *generator) Slice(t *types.Slice, name string) locations {
	loc := g.newGenLocation(t, name)

	et := t.Elem()
	typeName := g.TypeNameFor(et)
	eloc := g.Type(et, typeName)

	g.W(`func %s(s *serde.Serializer, x %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	g.W(`b = serde.SerializeSize(len(x), b)`)
	g.W(`for _, x := range x {`)
	g.serializeCallForLoc(eloc)
	g.W(`}`)
	g.W(`return b`)
	g.W(`}`)
	g.W(``)

	g.W(`func %s(d *serde.Deserializer, b []byte) (%s, []byte) {`, loc.deserializer.name, name)
	g.W(`d = serde.EnsureDeserializer(d)`)
	g.W(`n, b := serde.DeserializeSize(b)`)
	g.W(`var z %s`, name)
	g.W(`for i := 0; i < n; i++ {`)
	g.W(`var x %s`, typeName)
	g.deserializeCallForLoc(eloc)
	g.W(`z = append(z, x)`)
	g.W(`}`)
	g.W(`return z, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *generator) Named(t *types.Named, name string) locations {
	typeName := g.TypeNameFor(t.Obj().Type())
	return g.Type(t.Underlying(), typeName)
}

func (g *generator) Struct(t *types.Struct, name string) locations {
	loc := g.newGenLocation(t, name)

	// Depth-first search in the fields to generate serialization functions
	// of fields themsleves.
	n := t.NumFields()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		ft := f.Type()
		typeName := g.TypeNameFor(ft)
		g.Type(ft, typeName)
	}

	// Generate a new function to serialize this struct type.
	g.W(`func %s(s *serde.Serializer, x %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	// TODO: private fields
	for i := 0; i < n; i++ {
		f := t.Field(i)
		ft := f.Type()

		typeName := g.TypeNameFor(ft)
		floc := g.Type(ft, typeName)

		g.W(`{`)
		g.W(`x := x.%s`, f.Name())
		g.serializeCallForLoc(floc)
		g.W(`}`)
	}
	g.W(`return b`)
	g.W(`}`)
	g.W(``)

	g.W(`func %s(d *serde.Deserializer, b []byte) (%s, []byte) {`, loc.deserializer.name, name)
	g.W(`d = serde.EnsureDeserializer(d)`)
	g.W(`var z %s`, name)
	// TODO: private fields
	for i := 0; i < n; i++ {
		f := t.Field(i)
		ft := f.Type()

		typeName := g.TypeNameFor(ft)
		floc := g.Type(ft, typeName)

		g.W(`{`)
		g.W(`var x %s`, typeName)
		g.deserializeCallForLoc(floc)
		g.W(`z.%s = x`, f.Name())
		g.W(`}`)
	}
	g.W(`return z, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *generator) serializeCallForLoc(loc locations) {
	l := loc.serializer

	args := "s, x, b"
	if l.terminal {
		args = "x, b"
	}

	if l.pkg != "" {
		g.W(`b = %s.%s(%s)`, l.pkg, l.name, args)
	} else {
		g.W(`b = %s(%s)`, l.name, args)
	}
}

func (g *generator) deserializeCallForLoc(loc locations) {
	l := loc.deserializer

	args := "d, b"
	if l.terminal {
		args = "b"
	}

	if l.pkg != "" {
		g.W(`x, b = %s.%s(%s)`, l.pkg, l.name, args)
	} else {
		g.W(`x, b = %s(%s)`, l.name, args)
	}
}

func isInvalidChar(r rune) bool {
	valid := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r == '_')
	return !valid
}

// Generate, save, and return a new location for a type with generated
// serializers.
func (g *generator) newGenLocation(t types.Type, name string) locations {
	//TODO: check name collision
	if strings.ContainsFunc(name, isInvalidChar) {
		name = ""
	}
	if name == "" {
		name = fmt.Sprintf("gen%d", g.known.Len())
	}
	loc := locations{
		serializer: location{
			name:     "Serialize_" + name,
			terminal: false,
		},
		deserializer: location{
			name:     "Deserialize_" + name,
			terminal: false,
		},
	}
	g.setLocation(t, loc)
	return loc
}

func (g *generator) setLocation(t types.Type, loc locations) {
	prev := g.known.Set(t, loc)
	if prev != nil {
		panic(fmt.Errorf("trying to override known location"))
	}
}

func (g *generator) builtin(t types.Type, ser, des interface{}) locations {
	g.imports.Ensure("serde", "github.com/stealthrocket/coroutine/serde")
	nameof := func(x interface{}) string {
		if s, ok := x.(string); ok {
			return s
		}

		full := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
		return full[strings.LastIndexByte(full, '.')+1:]
	}
	l := locations{
		serializer: location{
			pkg:      "serde",
			name:     nameof(ser),
			terminal: true,
		},
		deserializer: location{
			pkg:      "serde",
			name:     nameof(des),
			terminal: true,
		},
	}
	g.setLocation(t, l)
	return l
}

func (g *generator) Basic(t *types.Basic, name string) locations {
	switch t.Kind() {
	case types.Invalid:
		panic("trying to generate serializer for invalid basic type")
	case types.String:
		return g.builtin(t, serde.SerializeString, serde.DeserializeString)
	case types.Bool:
		return g.builtin(t, serde.SerializeBool, serde.DeserializeBool)
	case types.Int:
		return g.builtin(t, serde.SerializeInt, serde.DeserializeInt)
	case types.Int64:
		return g.builtin(t, serde.SerializeInt64, serde.DeserializeInt64)
	case types.Int32:
		return g.builtin(t, serde.SerializeInt32, serde.DeserializeInt32)
	case types.Int16:
		return g.builtin(t, serde.SerializeInt16, serde.DeserializeInt16)
	case types.Int8:
		return g.builtin(t, serde.SerializeInt8, serde.DeserializeInt8)
	case types.Uint64:
		return g.builtin(t, serde.SerializeUint64, serde.DeserializeUint64)
	case types.Uint32:
		return g.builtin(t, serde.SerializeUint32, serde.DeserializeUint32)
	case types.Uint16:
		return g.builtin(t, serde.SerializeUint16, serde.DeserializeUint16)
	case types.Uint8:
		return g.builtin(t, serde.SerializeUint8, serde.DeserializeUint8)
	case types.Float32:
		return g.builtin(t, serde.SerializeFloat32, serde.DeserializeFloat32)
	case types.Float64:
		return g.builtin(t, serde.SerializeFloat64, serde.DeserializeFloat64)
	case types.Complex64:
		return g.builtin(t, serde.SerializeComplex64, serde.DeserializeComplex64)
	case types.Complex128:
		return g.builtin(t, serde.SerializeComplex128, serde.DeserializeComplex128)
	default:
		panic(fmt.Errorf("basic type kind %s not handled", basicKindString(t)))
	}
}

func (g *generator) TypeNameFor(t types.Type) string {
	return types.TypeString(t, types.RelativeTo(g.main.Types))
}

func (g *generator) get(t types.Type) (locations, bool) {
	loc := g.known.At(t)
	if loc == nil {
		return locations{}, false
	}
	return loc.(locations), true
}

type typedef struct {
	obj types.Object
	pkg *packages.Package
}

// TargetFile returns the path where a serder function should be generated for
// this type.
func (t typedef) TargetFile() string {
	pos := t.pkg.Fset.Position(t.obj.Pos())
	dir, file := filepath.Split(pos.Filename)

	i := strings.LastIndexByte(file, '.')
	if i == -1 {
		panic(fmt.Errorf("files does not end in .go: %s", file))
	}
	outFile := file[:i] + "_serde.go"
	return filepath.Join(dir, outFile)
}

var notype = typedef{}

func findTypeDef(name string, pkgs []*packages.Package) typedef {
	for _, pkg := range pkgs {
		for id, d := range pkg.TypesInfo.Defs {
			if id.Name == name {
				// TOOD: this probably need more checks.
				return typedef{obj: d, pkg: pkg}
			}
		}
	}
	return notype
}

func parse(patterns []string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
	}
	return packages.Load(cfg, patterns...)
}

func basicKindString(x *types.Basic) string {
	return [...]string{
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
	}[x.Kind()]
}
