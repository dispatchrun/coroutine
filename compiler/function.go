package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"slices"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
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

func collectFunctypes(p *packages.Package, name string, fn ast.Node, scope *funcscope, colors map[ast.Node]*types.Signature, functypes map[string]functype) {
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

	signature := functionTypeOf(fn)
	for _, fields := range []*ast.FieldList{signature.Params, signature.Results} {
		if fields != nil {
			for _, field := range fields.List {
				for _, name := range field.Names {
					scope.insert(name, field.Type)
				}
			}
		}
	}

	pre := func(cursor *astutil.Cursor) bool {
		switch n := cursor.Node().(type) {
		case *ast.Ident:
			observeIdent(n)

		case *ast.GenDecl:
			if n.Tok == token.VAR {
				for _, spec := range n.Specs {
					switch s := spec.(type) {
					case *ast.ValueSpec:
						for _, name := range s.Names {
							scope.insert(name, s.Type)
						}
					}
				}
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

	post := func(cursor *astutil.Cursor) bool {
		switch cursor.Node().(type) {
		case *ast.BlockStmt:
			scope = scope.outer
		}
		return true
	}

	astutil.Apply(functionBodyOf(fn), pre, post)

	functype := functype{
		signature: signature,
	}
	if len(freeVars) > 0 {
		fields := make([]*ast.Field, 1+len(freeVars))
		fields[0] = &ast.Field{
			Type:  ast.NewIdent("uintptr"),
			Names: []*ast.Ident{ast.NewIdent("F")},
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

			fields[i+1] = &ast.Field{
				Type:  fieldType,
				Names: []*ast.Ident{fieldName},
			}
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
			collectFunctypes(p, anonFuncName, anonFunc.node, anonFunc.scope, colors, functypes)
		}
	}
}

func generateFunctypes(p *packages.Package, f *ast.File, colors map[ast.Node]*types.Signature) *ast.File {
	functypes := map[string]functype{}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			scope := &funcscope{vars: map[string]*funcvar{}}
			name := p.PkgPath + "." + d.Name.Name
			collectFunctypes(p, name, d, scope, colors, functypes)
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

	gen := &ast.File{
		Name: ast.NewIdent(p.Name),
		Decls: []ast.Decl{
			&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					&ast.ImportSpec{
						Name: ast.NewIdent("_types"),
						Path: &ast.BasicLit{
							Kind:  token.STRING,
							Value: `"github.com/stealthrocket/coroutine/types"`,
						},
					},
				},
			},
			&ast.FuncDecl{
				Name: ast.NewIdent("init"),
				Type: &ast.FuncType{Params: new(ast.FieldList)},
				Body: init,
			},
		},
	}

	return addImports(p, gen)
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
