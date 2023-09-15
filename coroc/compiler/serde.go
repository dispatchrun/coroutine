package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

const publicSerdePackage = "github.com/stealthrocket/coroutine/serde"

// GenerateTypesInit searches pkg for types that serde can handle and append an
// init() function to the provided ast to register them on package load.
func generateTypesInit(fset *token.FileSet, gen *ast.File, pkg *packages.Package) error {
	// Prepare the imports map with the imports already in the AST.
	w := typesWalker{
		imports:   makeImportsMap(fset, gen),
		gen:       gen,
		seenTypes: make(map[string]struct{}),
		seenPkgs:  make(map[string]struct{}),
		from:      pkg.Types.Path(),
	}

	w.Walk(pkg)

	// Found no type.
	if len(w.types) == 0 {
		return nil
	}

	coropkg := ast.NewIdent(w.addImport(publicSerdePackage))

	sort.Slice(w.newimports, func(i, j int) bool {
		return w.newimports[i][0] < w.newimports[j][0]
	})

	for _, imp := range w.newimports {
		added := astutil.AddNamedImport(fset, gen, imp[0], imp[1])
		if !added {
			panic(fmt.Errorf(`import '%s' "%s" was supposed to be missing`, imp[0], imp[1]))
		}
	}

	fun := &ast.FuncDecl{
		Name: ast.NewIdent("init"),
		Type: &ast.FuncType{},
		Body: &ast.BlockStmt{},
	}

	sort.Strings(w.types)
	for _, t := range w.types {
		var typeExpr ast.Expr
		pkg, name, found := strings.Cut(t, ".")
		if found {
			// eg: coroutine.RegisterType[syscall.RtGenmsg]()
			typeExpr = &ast.SelectorExpr{
				X:   ast.NewIdent(pkg),
				Sel: ast.NewIdent(name),
			}
		} else {
			// eg: coroutine.RegisterType[MyStruct]()
			typeExpr = ast.NewIdent(pkg)
		}

		fun.Body.List = append(fun.Body.List, &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.IndexListExpr{
					X: &ast.SelectorExpr{
						X:   coropkg,
						Sel: ast.NewIdent("RegisterType"),
					},
					Indices: []ast.Expr{
						typeExpr,
					},
				},
			},
		})
	}

	gen.Decls = append(gen.Decls, fun)

	return nil
}

type typesWalker struct {
	imports   importsmap
	gen       *ast.File
	from      string
	seenTypes map[string]struct{}
	seenPkgs  map[string]struct{}

	// outputs
	types      []string
	newimports [][2]string // name, path
}

func (w *typesWalker) Walk(p *packages.Package) {
	if _, ok := w.seenPkgs[p.ID]; ok {
		return
	}
	w.seenPkgs[p.ID] = struct{}{}

	// Walk the type definitions in this package.
	for _, o := range p.TypesInfo.Defs {
		if o == nil {
			continue
		}
		t := o.Type()
		if !supported(w.from, t) {
			continue
		}

		w.add(t)
	}

	// Walk type instances from generics in this package.
	for _, i := range p.TypesInfo.Instances {
		if !supported(w.from, i.Type) {
			continue
		}
		w.add(i.Type)
	}

	// Recurse into imports.
	for _, i := range p.Imports {
		w.Walk(i)
	}
}

// Import the given package path, and return the actual name used in case of
// conflict.
func (w *typesWalker) addImport(path string) string {
	new, importname := w.imports.Add(path)
	if new {
		w.newimports = append(w.newimports, [2]string{importname, path})
	}
	return importname
}

func (w *typesWalker) add(t types.Type) {
	typename := types.TypeString(t, nil)
	if _, ok := w.seenTypes[typename]; ok {
		return
	}

	path, name, found := cutLast(typename, ".")
	if !found {
		name = path
	} else if path != w.from {
		importname := w.addImport(path)
		name = importname + "." + name
	}

	w.types = append(w.types, name)
	w.seenTypes[typename] = struct{}{}
}

func makeImportsMap(fset *token.FileSet, gen *ast.File) importsmap {
	m := importsmap{}

	importgroups := astutil.Imports(fset, gen)
	for _, g := range importgroups {
		for _, imp := range g {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				panic(fmt.Errorf("package in import not quoted: %w", err))
			}
			name := importSpecName(imp)
			m.Ensure(name, path)
		}
	}

	return m
}

func importSpecName(imp *ast.ImportSpec) string {
	if imp.Name != nil {
		return imp.Name.Name
	}
	// guess from the path
	path, err := strconv.Unquote(imp.Path.Value)
	if err != nil {
		panic(fmt.Errorf("package in import not quoted: %w", err))
	}
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return path
	}
	return path[lastSlash+1:]
}

func supported(from string, t types.Type) bool {
	return supportedType(t) && supportedImport(from, t)
}

func supportedImport(from string, t types.Type) bool {
	typename := types.TypeString(t, nil)
	path, name, found := cutLast(typename, ".")
	if !found {
		name = path
		path = ""
	}

	if !validPath(path) {
		return false
	}
	if !found {
		return true
	}
	if strings.Contains(path, "internal/") && !strings.HasPrefix(path, from+"/internal/") {
		return false
	}
	if !public(name) {
		return false
	}
	return true
}

func validPath(s string) bool {
	return !strings.ContainsAny(s, "[]<>{}* ")
}

func supportedType(t types.Type) bool {
	switch x := t.(type) {
	case *types.Signature:
		// don't know how to serialize functions
		return false
	case *types.Named:
		// uninstantiated type parameter
		if x.Origin() != t {
			return false
		}

		tp := x.TypeParams()
		if tp != nil {
			// Had type parameters at some point. need to check if
			// they are instantiated.
			if x.TypeArgs().Len() != tp.Len() {
				return false
			}
		}

		return supportedType(t.Underlying())
	case *types.Interface:
		// TODO: should this be relaxed?
		return false
	case *types.TypeParam:
		return false
	case *types.Chan:
		return false
	case *types.Pointer:
		return supportedType(x.Elem())
	case *types.Array:
		return supportedType(x.Elem())
	case *types.Slice:
		return supportedType(x.Elem())
	case *types.Map:
		return supportedType(x.Elem()) && supportedType(x.Key())
	case *types.Basic:
		switch x.Kind() {
		case types.UntypedBool,
			types.UntypedInt,
			types.UntypedRune,
			types.UntypedFloat,
			types.UntypedComplex,
			types.UntypedString,
			types.UntypedNil,
			types.Invalid:
			return false
		}
	}
	return true
}

type importsmap struct {
	byName map[string]string
	byPath map[string]string
}

// Add a new import and assign it an imported name, avoiding clashes. Return
// true if a new import was added.
func (m *importsmap) Add(p string) (bool, string) {
	name, ok := m.byPath[p]
	if ok {
		return false, name
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
		return true, name
	}

	panic("exhausted suffixes")
}

// Ensure path is imported, and its import name matches the one provided. Panics
// otherwise.
func (m *importsmap) Ensure(name, path string) {
	// Since the goal is to panic, it's ok to potentially add an import with
	// the wrong name.
	_, importname := m.Add(path)
	if importname != name {
		panic(fmt.Errorf("import package '%s' is imported as '%s'; expected '%s'", path, importname, name))
	}
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
