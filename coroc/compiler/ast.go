package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
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
			var idx *ast.Ident
			if hasIdx := s.Key != nil && !isUnderscore(s.Key); !hasIdx {
				idx = ast.NewIdent("_rangei" + strconv.Itoa(i))
				idx.Obj = &ast.Object{Kind: ast.Var, Name: idx.Name}
				info.Defs[idx] = types.NewVar(0, nil, idx.Name, types.Typ[types.Int])
			} else {
				idx = s.Key.(*ast.Ident)
			}

			x := ast.NewIdent("_rangex" + strconv.Itoa(i))
			x.Obj = &ast.Object{Kind: ast.Var, Name: x.Name}
			info.Defs[x] = types.NewVar(0, nil, x.Name, info.TypeOf(s.X))

			if hasVal := s.Value != nil && !isUnderscore(s.Value); hasVal {
				s.Body.List = append([]ast.Stmt{
					&ast.AssignStmt{Lhs: []ast.Expr{s.Value}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.IndexExpr{X: x, Index: idx}}},
				}, s.Body.List...)
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

func typeExpr(typ types.Type) ast.Expr {
	switch t := typ.(type) {
	case *types.Basic:
		return ast.NewIdent(t.String())
	case *types.Slice:
		return &ast.ArrayType{Elt: typeExpr(t.Elem())}
	case *types.Array:
		return &ast.ArrayType{
			Len: &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(t.Len(), 10)},
			Elt: typeExpr(t.Elem()),
		}
	case *types.Interface:
		if t.Empty() {
			return ast.NewIdent("any")
		}
	}
	panic(fmt.Sprintf("not implemented: %T", typ))
}
