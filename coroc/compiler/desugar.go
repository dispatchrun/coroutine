package compiler

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
)

// desugar recursively simplifies an AST. The goal is to hoist initialization
// statements out of branches and loops, so that when resuming a coroutine
// within that branch or loop the initialization can be skipped. Other types
// of desugaring may be required in the future.
//
// At this time, desugaring is performed *after* packages have been loaded,
// parsed and type-checked, which means that any AST transformations below
// that introduce temporary variables must also update the associated
// types.Info. If this gets unruly in the future, desugaring should be
// performed after parsing AST's but before type checking so that this is
// done by the type checker.
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
				info.Defs[idx] = types.NewVar(0, nil, idx.Name, types.Typ[types.Int])
			} else {
				idx = s.Key.(*ast.Ident)
			}

			x := ast.NewIdent("_rangex" + strconv.Itoa(i))
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
			// Rewrite `switch init; tag {}` to `init; switch tag {}`
			if init := s.Init; init != nil {
				s.Init = nil
				stmt = &ast.BlockStmt{List: []ast.Stmt{init, s}}
			}
		case *ast.TypeSwitchStmt:
			// Rewrite `switch init; assign {}` to `init; switch assign {}`
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
