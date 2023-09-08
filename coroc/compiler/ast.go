package compiler

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/packages"
)

// desugar recursively desugars an AST. The goal is to hoist initialization
// statements out of branches and loops, so that when resuming a coroutine
// within that branch or loop the initialization can be skipped. Other types
// of desugaring may be required in the future.
func desugar(tree ast.Node, info *types.Info) {
	ast.Inspect(tree, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.BlockStmt:
			n.List = desugar0(n.List, info)
		case *ast.CaseClause:
			n.Body = desugar0(n.Body, info)
		}
		return true
	})
}

func desugar0(stmts []ast.Stmt, info *types.Info) (desugared []ast.Stmt) {
	for i, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.IfStmt:
			// Recursively rewrite `else if init; cond {}` to `else { init; if cond {} }`
			curr := s
			for {
				elseIf, ok := curr.Else.(*ast.IfStmt)
				if !ok {
					break
				}
				if init := elseIf.Init; init != nil {
					elseIf.Init = nil
					curr.Else = &ast.BlockStmt{List: []ast.Stmt{init, elseIf}}
				}
				curr = elseIf
			}
			// Rewrite `if init; cond {}` to `{ init; if cond {} }`
			if init := s.Init; init != nil {
				s.Init = nil
				stmt = &ast.BlockStmt{List: []ast.Stmt{init, s}}
			}
		case *ast.ForStmt:
			// Rewrite `for init; cond; post {}` to `{ init; for ; cond; post {} }`
			if init := s.Init; init != nil {
				s.Init = nil
				stmt = &ast.BlockStmt{List: []ast.Stmt{init, s}}
			}
		case *ast.RangeStmt:
			// Rewrite for range loops over arrays/slices.
			// - `for range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
			// - `for _ := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
			// - `for _, _ := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
			// - `for i := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ {} }`
			// - `for i, _ := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ {} }`
			// - `for i, v := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ { v := _x[i]; ... } }`
			// - `for _, v := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ { v := _x[_i]; ... } }`
			hasIdx := s.Key != nil && !isUnderscore(s.Key)
			hasVal := s.Value != nil && !isUnderscore(s.Value)
			prefix := "_d" + strconv.Itoa(i)
			idx := s.Key
			if !hasIdx {
				idx = ast.NewIdent(prefix + "_i")
			}
			x := ast.NewIdent(prefix + "_x")
			body := s.Body
			if hasVal {
				body.List = append([]ast.Stmt{
					&ast.AssignStmt{Lhs: []ast.Expr{x}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.IndexExpr{X: x, Index: idx}}},
				}, body.List...)
			}
			stmt = &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{Lhs: []ast.Expr{x}, Tok: token.DEFINE, Rhs: []ast.Expr{s.X}},
					&ast.ForStmt{
						Init: &ast.AssignStmt{Lhs: []ast.Expr{idx}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
						Post: &ast.IncDecStmt{X: idx, Tok: token.INC},
						Cond: &ast.BinaryExpr{X: idx, Op: token.LSS, Y: &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{x}}},
						Body: s.Body,
					},
				},
			}
		case *ast.SwitchStmt:
			// Rewrite `switch init; cond {}` to `init; switch cond {}`
			if init := s.Init; init != nil {
				s.Init = nil
				stmt = &ast.BlockStmt{List: []ast.Stmt{init, s}}
			}
		}
		desugared = append(desugared, stmt)
	}
	return
}

func isUnderscore(e ast.Expr) bool {
	i, ok := e.(*ast.Ident)
	return ok && i.Name == "_"
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

// trackSpans assigns a non-zero monotonically increasing integer ID to each
// leaf statement in the tree using a post-order traversal, and then assigns
// a "span" to all statements in the tree which is equal to the half-open
// range of IDs seen in that subtree.
func trackSpans(stmt ast.Stmt) map[ast.Stmt]span {
	spans := map[ast.Stmt]span{}
	trackSpans0(stmt, spans, 1)
	return spans
}

func trackSpans0(stmt ast.Stmt, spans map[ast.Stmt]span, nextID int) int {
	startID := nextID
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		for _, child := range s.List {
			nextID = trackSpans0(child, spans, nextID)
		}
	case *ast.IfStmt:
		nextID = trackSpans0(s.Body, spans, nextID)
		if s.Else != nil {
			nextID = trackSpans0(s.Else, spans, nextID)
		}
	case *ast.ForStmt:
		nextID = trackSpans0(s.Body, spans, nextID)
	case *ast.SwitchStmt:
		nextID = trackSpans0(s.Body, spans, nextID)
	case *ast.CaseClause:
		for _, child := range s.Body {
			nextID = trackSpans0(child, spans, nextID)
		}
	default:
		nextID++ // leaf
	}
	spans[stmt] = span{startID, nextID}
	return nextID
}

// unnestBlocks recursively unnests blocks with just one statement.
func unnestBlocks(stmt ast.Stmt) ast.Stmt {
	for {
		s, ok := stmt.(*ast.BlockStmt)
		if !ok || len(s.List) != 1 {
			return stmt
		}
		stmt = s.List[0]
	}
}
