package compiler

import (
	"cmp"
	"go/ast"
	"go/token"
	"slices"
	"strconv"

	"golang.org/x/tools/go/ssa"
)

func generateFunctypes(pkg *ssa.Package) *ast.File {
	var names = make([]string, 0, len(pkg.Members))
	for name := range pkg.Members {
		names = append(names, name)
	}
	slices.Sort(names)

	var init ast.BlockStmt
	var path = pkg.Pkg.Path()
	for _, name := range names {
		if fn, ok := pkg.Members[name].(*ssa.Function); ok {
			generateFunctypesInit(path, &init, fn)
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

func generateFunctypesInit(path string, init *ast.BlockStmt, fn *ssa.Function) {
	if fn.TypeParams() != nil {
		return // ignore non-instantiated generic functions
	}

	init.List = append(init.List, &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.IndexListExpr{
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("_types"),
					Sel: ast.NewIdent("RegisterFunc"),
				},
				Indices: []ast.Expr{
					newFuncType(fn.Signature),
				},
			},
			Args: []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(path + "." + fn.Name()),
				},
			},
		},
	})

	anonFuncs := slices.Clone(fn.AnonFuncs)
	slices.SortFunc(anonFuncs, func(f1, f2 *ssa.Function) int {
		return cmp.Compare(f1.Name(), f2.Name())
	})

	for _, anonFunc := range anonFuncs {
		generateFunctypesInit(path, init, anonFunc)
	}
}
