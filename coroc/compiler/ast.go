package compiler

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// desugar recursively desugars a set of statements. The goal is to
// hoist initialization statements out of branches and loops, so that
// when resuming a coroutine within that branch or loop the
// initialization can be skipped. Other types of desugaring may be
// required in the future.
func desugar(stmts []ast.Stmt) (desugared []ast.Stmt) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.BlockStmt:
			s.List = desugar(s.List)
		case *ast.ForStmt:
			if s.Init != nil {
				desugared = append(desugared, s.Init)
				s.Init = nil
				desugared = append(desugared, s)
				continue
			}
		}
		desugared = append(desugared, stmt)
	}
	return
}

// scanYields searches for cases of coroutine.Yield[R,S] in a tree.
//
// It handles cases where the coroutine package was imported with an alias
// or with a dot import. It doesn't currently handle cases where the yield
// types are inferred. It only partially handles references to the yield
// function (e.g. a := coroutine.Yield[R,S]; a()); if the reference is taken
// within the tree then the yield and its types will be reported, however if
// the reference was taken outside the tree it will not be seen here.
func scanYields(p *packages.Package, tree ast.Node, fn func(types []ast.Expr) bool) {
	ast.Inspect(tree, func(node ast.Node) bool {
		indexListExpr, ok := node.(*ast.IndexListExpr)
		if !ok {
			return true
		}
		switch x := indexListExpr.X.(type) {
		case *ast.Ident: // Yield[R,S]
			if x.Name != coroutineYield {
				return true
			} else if uses, ok := p.TypesInfo.Uses[x]; !ok {
				return true
			} else if fn, ok := uses.(*types.Func); !ok {
				return true
			} else if pkg := fn.Pkg(); pkg == nil || pkg.Path() != coroutinePackage {
				return true
			}
		case *ast.SelectorExpr: // coroutine.Yield[R,S]
			if x.Sel.Name != coroutineYield {
				return true
			} else if selX, ok := x.X.(*ast.Ident); !ok {
				return true
			} else if uses, ok := p.TypesInfo.Uses[selX]; !ok {
				return true
			} else if pkg, ok := uses.(*types.PkgName); !ok || pkg.Imported().Path() != coroutinePackage {
				return true
			}
		default:
			return true
		}
		return fn(indexListExpr.Indices)
	})
}

type span struct{ start, end int }

// trackSpans assigns an integer ID to each leaf statement in the tree using
// a post-order traversal, and then assigns a "span" to all statements in the
// tree which is equal to the half-open range of IDs seen in that subtree.
func trackSpans(stmt ast.Stmt) map[ast.Stmt]span {
	spans := map[ast.Stmt]span{}
	trackSpans0(stmt, spans, 0)
	return spans
}

func trackSpans0(stmt ast.Stmt, spans map[ast.Stmt]span, nextID int) int {
	startID := nextID
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		for _, child := range s.List {
			nextID = trackSpans0(child, spans, nextID)
		}
	case *ast.ForStmt:
		nextID = trackSpans0(s.Body, spans, nextID)
	default:
		nextID++ // leaf
	}
	spans[stmt] = span{startID, nextID}
	return nextID
}
