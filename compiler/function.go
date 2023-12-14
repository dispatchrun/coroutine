package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"maps"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type functype struct {
	signature ast.Expr
	closure   ast.Expr
}

type funcscope struct {
	outer *funcscope
	vars  map[string]*funcvar
}

func (s *funcscope) fork() *funcscope {
	return &funcscope{outer: s.outer, vars: maps.Clone(s.vars)}
}

func (s *funcscope) insert(name *ast.Ident, typ ast.Expr) {
	if name.Name != "_" {
		s.vars[name.Name] = &funcvar{scope: s, name: name, typ: typ}
	}
}

func (s *funcscope) lookup(name string) *funcvar {
	if name == "_" {
		return nil
	}
	if s == nil {
		return nil
	}
	if v, ok := s.vars[name]; ok {
		return v
	}
	return s.outer.lookup(name)
}

func (s *funcscope) in(other *funcscope) bool {
	return s == other || (s != nil && s.outer.in(other))
}

type funcvar struct {
	scope *funcscope
	name  *ast.Ident
	typ   ast.Expr
}

func collectFunctypes(p *packages.Package, name string, fn ast.Node, scope *funcscope, colors map[ast.Node]*types.Signature, functypes map[string]functype, g *genericInstance) {
	type function struct {
		node  ast.Node
		scope *funcscope
	}

	var funcScope = scope
	var freeNames = map[string]struct{}{}
	var freeVars []*funcvar
	var anonFuncs []function

	observeIdent := func(ident *ast.Ident) *funcvar {
		v := scope.lookup(ident.Name)
		if v != nil {
			if !v.scope.in(funcScope) {
				if _, seen := freeNames[ident.Name]; !seen {
					freeNames[ident.Name] = struct{}{}
					freeVars = append(freeVars, v)
				}
			}
		}
		return v
	}

	signature := copyFunctionType(functionTypeOf(fn))
	signature.TypeParams = nil

	recv := copyFieldList(functionRecvOf(fn))
	for _, fields := range []*ast.FieldList{recv, signature.Params, signature.Results} {
		if fields != nil {
			for _, field := range fields.List {
				for _, name := range field.Names {
					typ := p.TypesInfo.TypeOf(name)
					if g != nil {
						if instanceType, ok := g.typeOfParam(typ); ok {
							typ = instanceType
						}
					}
					if typ != nil {
						_, ellipsis := field.Type.(*ast.Ellipsis)
						field.Type = typeExpr(p, typ)
						if a, ok := field.Type.(*ast.ArrayType); ok && a.Len == nil && ellipsis {
							field.Type = &ast.Ellipsis{Elt: a.Elt}
						}
					}
					scope.insert(name, field.Type)
				}
			}
		}
	}

	var inspect func(ast.Node) bool
	inspect = func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.Ident:
			observeIdent(n)

		case *ast.SelectorExpr:
			ast.Inspect(n.X, inspect)
			return false

		case *ast.GenDecl:
			if n.Tok == token.VAR {
				for _, spec := range n.Specs {
					switch s := spec.(type) {
					case *ast.ValueSpec:
						for _, name := range s.Names {
							typ := p.TypesInfo.TypeOf(name)
							if g != nil {
								if instanceType, ok := g.typeOfParam(typ); ok {
									typ = instanceType
								}
							}
							if typ == nil {
								scope.insert(name, s.Type)
							} else {
								scope.insert(name, typeExpr(p, typ))
							}
						}
					}
				}
				return false
			}

		case *ast.FuncLit:
			anonFuncs = append(anonFuncs, function{
				node:  n,
				scope: scope.fork(),
			})
			return false

		case *ast.BlockStmt:
			scope = &funcscope{outer: scope, vars: map[string]*funcvar{}}
		}
		return true
	}

	astutil.Apply(functionBodyOf(fn),
		func(cursor *astutil.Cursor) bool {
			return inspect(cursor.Node())
		},
		func(cursor *astutil.Cursor) bool {
			switch cursor.Node().(type) {
			case *ast.BlockStmt:
				scope = scope.outer
			}
			return true
		},
	)

	functype := functype{
		signature: signature,
	}
	if len(freeVars) > 0 {
		fields := []*ast.Field{
			{
				Type:  ast.NewIdent("uintptr"),
				Names: []*ast.Ident{ast.NewIdent("F")},
			},
		}
		if g != nil {
			// Append a field for the dictionary.
			fields = append(fields, &ast.Field{
				Type:  ast.NewIdent("uintptr"),
				Names: []*ast.Ident{ast.NewIdent("D")},
			})
		}
		for i, freeVar := range freeVars {
			fieldName := ast.NewIdent(fmt.Sprintf("X%d", i))
			fieldType := freeVar.typ

			// The Go compiler uses a more advanced mechanism to determine if a
			// free variable should be captured by pointer or by value: it looks
			// at whether the variable is reassigned, its address taken, and if
			// it is less than 128 bytes in size.
			//
			// We know that our closures will only capture pointers to stack
			// frames which are never reassigned nor have their addresses taken,
			// and pointers will be less than 128 bytes on all platforms, which
			// means that the stack frame pointer is always captured by value.

			fields = append(fields, &ast.Field{
				Type:  fieldType,
				Names: []*ast.Ident{fieldName},
			})
		}
		functype.closure = &ast.StructType{
			Fields: &ast.FieldList{List: fields},
		}
	}
	functypes[name] = functype

	if len(anonFuncs) > 0 {
		index := 0
		// Colored functions (those rewritten into coroutines) have a
		// deferred anonymous function injected at the beginning to perform
		// stack unwinding, which takes the ".func1" name.
		_, colored := colors[fn]
		if colored {
			index = 1
		}

		for i, anonFunc := range anonFuncs[index:] {
			anonFuncName := anonFuncLinkName(name, index+i+1)
			collectFunctypes(p, anonFuncName, anonFunc.node, anonFunc.scope, colors, functypes, g)
		}
	}
}

func packagePath(p *packages.Package) string {
	if p.Name == "main" {
		return "main"
	} else {
		return p.PkgPath
	}
}

func functionPath(p *packages.Package, f *ast.FuncDecl) string {
	var b strings.Builder
	b.WriteString(packagePath(p))
	if f.Recv != nil {
		signature := p.TypesInfo.Defs[f.Name].Type().(*types.Signature)
		recvType := signature.Recv().Type()
		isptr := false
		if ptr, ok := recvType.(*types.Pointer); ok {
			recvType = ptr.Elem()
			isptr = true
		}
		b.WriteByte('.')
		if isptr {
			b.WriteString("(*")
		}
		switch t := recvType.(type) {
		case *types.Named:
			b.WriteString(t.Obj().Name())
		default:
			panic(fmt.Sprintf("not implemented: %T", t))
		}
		if isptr {
			b.WriteByte(')')
		}
	}
	b.WriteByte('.')
	b.WriteString(f.Name.Name)
	return b.String()
}

func (c *compiler) generateFunctypes(p *packages.Package, f *ast.File, colors map[ast.Node]*types.Signature) {
	functypes := map[string]functype{}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			obj := p.TypesInfo.ObjectOf(d.Name).(*types.Func)
			fn := c.prog.FuncValue(obj)
			if fn.TypeParams() != nil {
				instances := c.generics[fn]
				if len(instances) == 0 {
					// This can occur when a generic function is never instantiated/used,
					// or when it's instantiated in a package not known to the compiler.
					log.Printf("warning: cannot register runtime type information for generic function %s", fn)
					continue
				}
				for _, instance := range instances {
					g := newGenericInstance(fn, instance)
					if g.partial() {
						// Skip instances where not all type params have concrete types.
						// I'm not sure why these are generated in the SSA program.
						continue
					}
					scope := &funcscope{vars: map[string]*funcvar{}}
					name := g.gcshapePath()
					collectFunctypes(p, name, d, scope, colors, functypes, g)
				}
			} else {
				scope := &funcscope{vars: map[string]*funcvar{}}
				name := functionPath(p, d)
				collectFunctypes(p, name, d, scope, colors, functypes, nil)
			}
		}
	}

	names := make([]string, 0, len(functypes))
	for name := range functypes {
		names = append(names, name)
	}
	slices.Sort(names)

	init := new(ast.BlockStmt)
	for _, name := range names {
		ft := functypes[name]

		var register ast.Expr
		if ft.closure == nil {
			register = &ast.IndexListExpr{
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("_types"),
					Sel: ast.NewIdent("RegisterFunc"),
				},
				Indices: []ast.Expr{ft.signature},
			}
		} else {
			register = &ast.IndexListExpr{
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("_types"),
					Sel: ast.NewIdent("RegisterClosure"),
				},
				Indices: []ast.Expr{ft.signature, ft.closure},
			}
		}

		init.List = append(init.List, &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: register,
				Args: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: strconv.Quote(name),
					},
				},
			},
		})
	}

	if len(init.List) > 0 {
		astutil.AddNamedImport(nil, f, "_types", "github.com/stealthrocket/coroutine/types")

		f.Decls = append(f.Decls,
			&ast.FuncDecl{
				Name: ast.NewIdent("init"),
				Type: &ast.FuncType{Params: new(ast.FieldList)},
				Body: init,
			})
	}
}

// This function computes the name that the linker gives to anonymous functions,
// using the base name of their parent function and appending ".func<index>".
//
// The function works with multiple levels of nesting as each level adds another
// ".func<index>" suffix, with the index being local to the parent scope.
func anonFuncLinkName(base string, index int) string {
	return fmt.Sprintf("%s.func%d", base, index)
}

func functionTypeOf(fn ast.Node) *ast.FuncType {
	switch f := fn.(type) {
	case *ast.FuncDecl:
		return f.Type
	case *ast.FuncLit:
		return f.Type
	default:
		panic("node is neither *ast.FuncDecl or *ast.FuncLit")
	}
}

func functionSignatureOf(p *packages.Package, fn ast.Node) *types.Signature {
	switch f := fn.(type) {
	case *ast.FuncDecl:
		return p.TypesInfo.Defs[f.Name].Type().(*types.Signature)
	case *ast.FuncLit:
		return p.TypesInfo.TypeOf(f).(*types.Signature)
	default:
		panic("node is neither *ast.FuncDecl or *ast.FuncLit")
	}
}

func functionRecvOf(fn ast.Node) *ast.FieldList {
	switch f := fn.(type) {
	case *ast.FuncDecl:
		return f.Recv
	case *ast.FuncLit:
		return nil
	default:
		panic("node is neither *ast.FuncDecl or *ast.FuncLit")
	}
}

func copyFunctionType(f *ast.FuncType) *ast.FuncType {
	return &ast.FuncType{
		TypeParams: copyFieldList(f.TypeParams),
		Params:     copyFieldList(f.Params),
		Results:    copyFieldList(f.Results),
	}
}

func copyFieldList(f *ast.FieldList) *ast.FieldList {
	if f == nil {
		return nil
	}
	list := make([]*ast.Field, len(f.List))
	for i := range f.List {
		list[i] = copyField(f.List[i])
	}
	return &ast.FieldList{List: list}
}

func copyField(f *ast.Field) *ast.Field {
	c := *f
	return &c
}

func functionBodyOf(fn ast.Node) *ast.BlockStmt {
	switch f := fn.(type) {
	case *ast.FuncDecl:
		return f.Body
	case *ast.FuncLit:
		return f.Body
	default:
		panic("node is neither *ast.FuncDecl or *ast.FuncLit")
	}
}

type genericInstance struct {
	origin   *ssa.Function
	instance *ssa.Function

	recvPtr  bool
	recvType *types.Named

	types map[types.Type]types.Type
}

func newGenericInstance(origin, instance *ssa.Function) *genericInstance {
	g := &genericInstance{origin: origin, instance: instance}

	if recv := g.instance.Signature.Recv(); recv != nil {
		switch t := recv.Type().(type) {
		case *types.Pointer:
			g.recvPtr = true
			switch pt := t.Elem().(type) {
			case *types.Named:
				g.recvType = pt
			default:
				panic(fmt.Sprintf("not implemented: %T", t))
			}

		case *types.Named:
			g.recvType = t
		default:
			panic(fmt.Sprintf("not implemented: %T", t))
		}
	}

	g.types = map[types.Type]types.Type{}
	if g.recvType != nil {
		g.scanRecvTypeArgs(func(p *types.TypeParam, _ int, arg types.Type) {
			g.types[p.Obj().Type()] = arg
		})
	}
	g.scanTypeArgs(func(p *types.TypeParam, _ int, arg types.Type) {
		g.types[p.Obj().Type()] = arg
	})

	return g
}

func (g *genericInstance) typeOfParam(t types.Type) (types.Type, bool) {
	v, ok := g.types[t]
	return v, ok
}

func (g *genericInstance) partial() bool {
	sig := g.instance.Signature
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		if _, ok := params.At(i).Type().(*types.TypeParam); ok {
			return true
		}
	}
	results := sig.Results()
	for i := 0; i < results.Len(); i++ {
		if _, ok := results.At(i).Type().(*types.TypeParam); ok {
			return true
		}
	}
	return false
}

func (g *genericInstance) scanRecvTypeArgs(fn func(*types.TypeParam, int, types.Type)) {
	typeParams := g.instance.Signature.RecvTypeParams()
	typeArgs := g.recvType.TypeArgs()
	for i := 0; i < typeArgs.Len(); i++ {
		arg := typeArgs.At(i)
		param := typeParams.At(i)

		fn(param, i, arg)
	}
}

func (g *genericInstance) scanTypeArgs(fn func(*types.TypeParam, int, types.Type)) {
	params := g.origin.TypeParams()
	args := g.instance.TypeArgs()

	for i := 0; i < params.Len(); i++ {
		fn(params.At(i), i, args[i])
	}
}

func (g *genericInstance) gcshapePath() string {
	var path strings.Builder

	path.WriteString(g.origin.Pkg.Pkg.Path())

	if g.recvType != nil {
		path.WriteByte('.')
		if g.recvPtr {
			path.WriteString("(*")
		}
		path.WriteString(g.recvType.Obj().Name())

		if g.recvType.TypeParams() != nil {
			path.WriteByte('[')
			g.scanRecvTypeArgs(func(_ *types.TypeParam, i int, arg types.Type) {
				if i > 0 {
					path.WriteString(",")
				}
				writeGoShape(&path, arg)
			})
			path.WriteByte(']')
		}

		if g.recvPtr {
			path.WriteByte(')')
		}
	}

	path.WriteByte('.')
	path.WriteString(g.instance.Object().(*types.Func).Name())

	if g.origin.Signature.TypeParams() != nil {
		path.WriteByte('[')
		g.scanTypeArgs(func(_ *types.TypeParam, i int, arg types.Type) {
			if i > 0 {
				path.WriteString(",")
			}
			writeGoShape(&path, arg)
		})
		path.WriteByte(']')
	}

	return path.String()
}

func writeGoShape(b *strings.Builder, tt types.Type) {
	b.WriteString("go.shape.")

	switch t := tt.Underlying().(type) {
	case *types.Basic:
		b.WriteString(t.Name())
	case *types.Pointer:
		// All pointers resolve to *uint8.
		b.WriteString("*uint8")
	case *types.Interface:
		if t.Empty() {
			b.WriteString("interface{}")
		} else {
			panic(fmt.Sprintf("not implemented: %#v (%T)", tt, t))
		}
	default:
		panic(fmt.Sprintf("not implemented: %#v (%T)", tt, t))
	}
}
