package compiler

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

const (
	coroutinePackage = "github.com/stealthrocket/coroutine"
	coroutineYield   = "Yield"
)

func ScanYields(p *packages.Package, n ast.Node, fn func(types []ast.Expr) bool) {
	ast.Inspect(n, func(node ast.Node) bool {
		if indexListExpr, ok := node.(*ast.IndexListExpr); ok {
			if yieldTypes, ok := unpackYield(p, indexListExpr); ok {
				if !fn(yieldTypes) {
					return false
				}
			}
		}
		// TODO: handle cases where the yield types are inferred
		return true
	})
}

func unpackYield(p *packages.Package, indexListExpr *ast.IndexListExpr) ([]ast.Expr, bool) {
	switch x := indexListExpr.X.(type) {
	case *ast.Ident:
		if x.Name != coroutineYield {
			return nil, false
		}
		uses, ok := p.TypesInfo.Uses[x]
		if !ok {
			return nil, false
		}
		if x.Obj != nil {
			return nil, false // shadowed
		}
		fn, ok := uses.(*types.Func)
		if !ok {
			return nil, false
		}
		pkg := fn.Pkg()
		if pkg == nil || pkg.Path() != coroutinePackage {
			return nil, false
		}
	case *ast.SelectorExpr:
		if x.Sel.Name != coroutineYield {
			return nil, false
		}
		selX, ok := x.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		if selX.Obj != nil {
			return nil, false // shadowed
		}
		uses, ok := p.TypesInfo.Uses[selX]
		if !ok {
			return nil, false
		}
		pkg, ok := uses.(*types.PkgName)
		if !ok || pkg.Imported().Path() != coroutinePackage {
			return nil, false
		}
	default:
		return nil, false
	}
	return indexListExpr.Indices, true
}
