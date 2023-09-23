package compiler

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strconv"

	"golang.org/x/tools/go/ssa"
)

func generateFunctypes(pkg *ssa.Package, colors functionColors) *ast.File {
	var names = make([]string, 0, len(pkg.Members))
	for name := range pkg.Members {
		names = append(names, name)
	}
	slices.Sort(names)

	var init ast.BlockStmt
	for _, name := range names {
		if fn, ok := pkg.Members[name].(*ssa.Function); ok {
			name := pkg.Pkg.Path() + "." + fn.Name()
			generateFunctypesInit(pkg, fn, &init, name, colors)
		}
	}

	return &ast.File{
		Name: ast.NewIdent(pkg.Pkg.Name()),
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
				Type: &ast.FuncType{},
				Body: &init,
			},
		},
	}
}

func generateFunctypesInit(pkg *ssa.Package, fn *ssa.Function, init *ast.BlockStmt, name string, colors functionColors) {
	if fn.TypeParams() != nil {
		return // ignore non-instantiated generic functions
	}
	var register ast.Expr
	if len(fn.FreeVars) == 0 {
		register = &ast.IndexListExpr{
			X: &ast.SelectorExpr{
				X:   ast.NewIdent("_types"),
				Sel: ast.NewIdent("RegisterFunc"),
			},
			Indices: []ast.Expr{
				newFuncType(pkg.Pkg, fn.Signature),
			},
		}
	} else {
		fields := make([]*ast.Field, 1+len(fn.FreeVars))
		// first field is the function address (uintptr)
		fields[0] = &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("_")},
			Type:  ast.NewIdent("uintptr"),
		}

		for i, freeVar := range fn.FreeVars {
			fields[i+1] = &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(freeVar.Name())},
				Type:  typeExpr(pkg.Pkg, freeVar.Type()),
			}
		}

		register = &ast.IndexListExpr{
			X: &ast.SelectorExpr{
				X:   ast.NewIdent("_types"),
				Sel: ast.NewIdent("RegisterClosure"),
			},
			Indices: []ast.Expr{
				newFuncType(pkg.Pkg, fn.Signature),
				&ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
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

	anonFuncs := slices.Clone(fn.AnonFuncs)
	slices.SortFunc(anonFuncs, func(f1, f2 *ssa.Function) int {
		return cmp.Compare(f1.Name(), f2.Name())
	})

	index := 0
	// Colored functions (those rewritten into coroutines) have a
	// deferred anonymous function injected at the beginning to perform
	// stack unwinding, which takes the ".func1" name.
	_, colored := colors[fn]
	if colored {
		index++
	}

	for _, anonFunc := range anonFuncs {
		index++
		anonFuncName := anonFuncLinkName(name, index)
		generateFunctypesInit(pkg, anonFunc, init, anonFuncName, colors)
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

// This function returns true if a function body is composed of at most one
// expression.
func functionBodyIsExpr(body *ast.BlockStmt) bool {
	if len(body.List) == 0 {
		return true
	}
	if len(body.List) == 1 {
		if _, isExpr := body.List[0].(*ast.ExprStmt); isExpr {
			return true
		}
	}
	return false
}
