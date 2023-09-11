package serde

import (
	"fmt"
	"go/types"
	"io"
	"log/slog"
	"math"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type location struct {
	pkg  string
	name string
}

func (l location) FullName() string {
	if l.pkg != "" {
		return fmt.Sprintf("%s.%s", l.pkg, l.name)
	}
	return l.name
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

type Generator struct {
	// Build tags to decorate the output
	tags []string
	// Type of the serde.Serializable interface.
	serializable *types.Interface
	// Map[types.Type] -> locations to track the types that already have
	// their serialization functions emitted.
	known typeutil.Map
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
	serializable := FindTypeDef("github.com/stealthrocket/coroutine/serde", "Serializable", pkgs)
	if serializable == Notype {
		panic("could not find built-in Serializable interface; make sure coroutine/serde is in pkgs")
	}
	serializableIface := serializable.Obj.Type().(*types.Named).Underlying().(*types.Interface)
	slog.Debug("found Serializable interface", "type", serializableIface)
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

func (g *Generator) GenType(t types.Type) {
	g.Type(t)
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
	// Generate a reflect.Type <> ID mapping.
	s := map[string]types.Type{}
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
			s[name] = t
		}

		for _, i := range p.Imports {
			findTypes(i)
		}
	}

	for _, pkg := range pkgs {
		findTypes(pkg)
	}

	reflectName := g.imports.Add("reflect")
	g.W(`func init() {`)
	g.W(`var t %s.Type`, reflectName)
	//	g.W(`serde.RegisterTypes(`)
	for k, t := range s {
		g.W(`{`)
		g.W(`var x %s`, k)
		g.W(`t = reflect.TypeOf(x)`)

		locs := g.known.At(t)
		if locs == nil {
			g.W(`serde.RegisterType(t)`)
		} else {
			x := locs.(locations)
			ser := x.serializer.FullName()
			des := x.deserializer.FullName()
			g.W(`sw := func(s *serde.Serializer, x any, b []byte) []byte {`)
			g.W(`return %s(s, x.(%s), b)`, ser, k)
			g.W(`}`)
			g.W(`dw := func(d *serde.Deserializer, b []byte) (any, []byte) {`)
			g.W(`return %s(d, b)`, des)
			g.W(`}`)
			g.W(`serde.RegisterTypeWithCodec(t, sw, dw)`)
		}
		g.W(`}`)
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

func (g *Generator) Type(t types.Type) locations {
	// Hit the cache first.
	if loc, ok := g.get(t); ok {
		return loc
	}

	if types.Implements(t, g.serializable) {
		return g.Serializable(t)
	}

	if types.Implements(types.NewPointer(t), g.serializable) {
		return g.SerializableToPtr(t)
	}

	switch x := t.(type) {
	case *types.Basic:
		return g.Basic(x)
	case *types.Struct:
		return g.Struct(x)
	case *types.Named:
		return g.Named(x)
	case *types.Slice:
		return g.Slice(x)
	case *types.Pointer:
		return g.Pointer(x)
	case *types.Map:
		return g.Map(x)
	case *types.Interface:
		return g.Interface(x)
	default:
		panic(fmt.Errorf("type generator not implemented: %s (%T)", t, t))
	}
}

func (g *Generator) Interface(t *types.Interface) locations {
	g.imports.Ensure("serde", "github.com/stealthrocket/coroutine/serde")
	l := locations{
		serializer: location{
			pkg:  "serde",
			name: nameof(SerializeInterface),
		},
		deserializer: location{
			pkg:  "serde",
			name: nameof(DeserializeInterface),
		},
	}
	g.setLocation(t, l)
	return l
}

func (g *Generator) Map(t *types.Map) locations {
	name := g.TypeNameFor(t)
	loc := g.newGenLocation(t, name)

	kt := t.Key()
	kname := g.TypeNameFor(kt)
	kloc := g.Type(kt)

	vt := t.Elem()
	vname := g.TypeNameFor(vt)
	vloc := g.Type(vt)

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

func (g *Generator) Pointer(t *types.Pointer) locations {
	unsafeName := g.imports.Add("unsafe")
	name := g.TypeNameFor(t)
	loc := g.newGenLocation(t, name)

	pt := t.Elem()
	ploc := g.Type(pt)
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

func (g *Generator) Serializable(t types.Type) locations {
	slog.Debug("using Serializable interface", "type", t)
	return g.builtin(t, "SerializeSerializable", "DeserializeSerializable")
}

func (g *Generator) SerializableToPtr(t types.Type) locations {
	slog.Debug("using pointer to Serializable interface", "type", t)
	// t is not Serializable, but *t is.
	name := g.TypeNameFor(t)
	loc := g.newGenLocation(t, name)

	// location for the pointer type
	ploc := g.Type(types.NewPointer(t))

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
	g.W(`b = serde.DeserializeSerializable(d, x, b)`)
	g.W(`return z, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *Generator) Slice(t *types.Slice) locations {
	name := g.TypeNameFor(t)
	loc := g.newGenLocation(t, name)

	et := t.Elem()
	ename := g.TypeNameFor(et)
	eloc := g.Type(et)

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
	g.W(`var x %s`, ename)
	g.deserializeCallForLoc(eloc)
	g.W(`z = append(z, x)`)
	g.W(`}`)
	g.W(`return z, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *Generator) Named(t *types.Named) locations {
	name := g.TypeNameFor(t.Obj().Type())
	loc := g.newGenLocation(t, name)

	u := t.Underlying()
	utype := g.Type(u)
	uname := g.TypeNameFor(u)

	g.W(`func %s(s *serde.Serializer, z %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	g.W(`x := (%s)(z)`, uname)
	g.serializeCallForLoc(utype)
	g.W(`return b`)
	g.W(`}`)

	g.W(`func %s(d *serde.Deserializer, b []byte) (%s, []byte) {`, loc.deserializer.name, name)
	g.W(`d = serde.EnsureDeserializer(d)`)
	g.W(`var x %s`, uname)
	g.deserializeCallForLoc(utype)
	g.W(`return (%s)(x), b`, name)
	g.W(`}`)

	return loc
}

func (g *Generator) Struct(t *types.Struct) locations {
	name := g.TypeNameFor(t)
	loc := g.newGenLocation(t, name)

	// Depth-first search in the fields to generate serialization functions
	// of fields themsleves.
	n := t.NumFields()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		ft := f.Type()
		g.Type(ft)
	}

	// Generate a new function to serialize this struct type.
	g.W(`func %s(s *serde.Serializer, x %s, b []byte) []byte {`, loc.serializer.name, name)
	g.W(`s = serde.EnsureSerializer(s)`)
	// TODO: private fields
	for i := 0; i < n; i++ {
		f := t.Field(i)
		ft := f.Type()
		floc := g.Type(ft)

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

		ename := g.TypeNameFor(ft)
		floc := g.Type(ft)

		g.W(`{`)
		g.W(`var x %s`, ename)
		g.deserializeCallForLoc(floc)
		g.W(`z.%s = x`, f.Name())
		g.W(`}`)
	}
	g.W(`return z, b`)
	g.W(`}`)
	g.W(``)

	return loc
}

func (g *Generator) serializeCallForLoc(loc locations) {
	l := loc.serializer

	args := "s, x, b"

	if l.pkg != "" {
		g.W(`b = %s.%s(%s)`, l.pkg, l.name, args)
	} else {
		g.W(`b = %s(%s)`, l.name, args)
	}
}

func (g *Generator) deserializeCallForLoc(loc locations) {
	l := loc.deserializer

	args := "d, b"

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
func (g *Generator) newGenLocation(t types.Type, name string) locations {
	slog.Debug("generated serde", "name", name, "type", t)
	//TODO: check name collision
	if strings.ContainsFunc(name, isInvalidChar) {
		name = ""
	}
	if name == "" {
		name = fmt.Sprintf("gen%d", g.known.Len())
	}
	loc := locations{
		serializer: location{
			name: "Serialize_" + name,
		},
		deserializer: location{
			name: "Deserialize_" + name,
		},
	}
	g.setLocation(t, loc)
	return loc
}

func (g *Generator) setLocation(t types.Type, loc locations) {
	prev := g.known.Set(t, loc)
	if prev != nil {
		panic(fmt.Errorf("trying to override known location"))
	}
}

func nameof(x interface{}) string {
	if s, ok := x.(string); ok {
		return s
	}

	full := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
	return full[strings.LastIndexByte(full, '.')+1:]
}

func (g *Generator) builtin(t types.Type, ser, des interface{}) locations {
	g.imports.Ensure("serde", "github.com/stealthrocket/coroutine/serde")
	l := locations{
		serializer: location{
			pkg:  "serde",
			name: nameof(ser),
		},
		deserializer: location{
			pkg:  "serde",
			name: nameof(des),
		},
	}
	g.setLocation(t, l)
	return l
}

func (g *Generator) Basic(t *types.Basic) locations {
	switch t.Kind() {
	case types.Invalid:
		panic("trying to generate serializer for invalid basic type")
	case types.String:
		return g.builtin(t, SerializeString, DeserializeString)
	case types.Bool:
		return g.builtin(t, SerializeBool, DeserializeBool)
	case types.Int:
		return g.builtin(t, SerializeInt, DeserializeInt)
	case types.Int64:
		return g.builtin(t, SerializeInt64, DeserializeInt64)
	case types.Int32:
		return g.builtin(t, SerializeInt32, DeserializeInt32)
	case types.Int16:
		return g.builtin(t, SerializeInt16, DeserializeInt16)
	case types.Int8:
		return g.builtin(t, SerializeInt8, DeserializeInt8)
	case types.Uint64:
		return g.builtin(t, SerializeUint64, DeserializeUint64)
	case types.Uint32:
		return g.builtin(t, SerializeUint32, DeserializeUint32)
	case types.Uint16:
		return g.builtin(t, SerializeUint16, DeserializeUint16)
	case types.Uint8:
		return g.builtin(t, SerializeUint8, DeserializeUint8)
	case types.Float32:
		return g.builtin(t, SerializeFloat32, DeserializeFloat32)
	case types.Float64:
		return g.builtin(t, SerializeFloat64, DeserializeFloat64)
	case types.Complex64:
		return g.builtin(t, SerializeComplex64, DeserializeComplex64)
	case types.Complex128:
		return g.builtin(t, SerializeComplex128, DeserializeComplex128)
	default:
		panic(fmt.Errorf("basic type kind %s not handled", basicKindString(t)))
	}
}

func (g *Generator) TypeNameFor(t types.Type) string {
	return types.TypeString(t, types.RelativeTo(g.main.Types))
}

func (g *Generator) get(t types.Type) (locations, bool) {
	loc := g.known.At(t)
	if loc == nil {
		return locations{}, false
	}
	return loc.(locations), true
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

type Typedef struct {
	Obj types.Object
	Pkg *packages.Package
}

// TargetFile returns the path where a serder function should be generated for
// this type.
func (t Typedef) TargetFile() string {
	pos := t.Pkg.Fset.Position(t.Obj.Pos())
	dir, file := filepath.Split(pos.Filename)

	i := strings.LastIndexByte(file, '.')
	if i == -1 {
		panic(fmt.Errorf("files does not end in .go: %s", file))
	}
	noext := file[:i]

	parts := strings.Split(noext, "_")

	fmt.Println(parts)
	parts = append(parts, "")
	copy(parts[2:], parts[1:])
	parts[1] = "serde"

	outFile := strings.Join(parts, "_") + ".go"
	return filepath.Join(dir, outFile)
}

var Notype = Typedef{}

func FindTypeDef(inpkg string, name string, pkgs []*packages.Package) Typedef {
	for _, pkg := range pkgs {
		if inpkg != "" && pkg.PkgPath != inpkg {
			continue
		}
		for id, d := range pkg.TypesInfo.Defs {
			if id.Name == name {
				// TOOD: this probably need more checks.
				return Typedef{Obj: d, Pkg: pkg}
			}
		}
	}
	return Notype
}
