package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
)

// desugar recursively replaces sugared AST nodes with simpler constructs.
//
// A goal is to hoist initialization statements out of branches and loops,
// so that when resuming a coroutine within that branch or loop the
// initialization can be skipped. Another goal is to hoist statements out
// of places where only one statement is valid, so that the statement can
// be decomposed as necessary.
//
// Implicit branch targets (e.g. via break/continue) are made explicit using
// labels so that the desugaring pass (and other compilation passes) are able
// to both decompose and introduce control flow.
//
// The desugaring pass works at the statement level (ast.Stmt) and does not
// consider expressions (ast.Expr). This means that the pass does not
// recurse into expressions that may contain statements. At this time, only
// one type of ast.Expr has nested statements, which is the function literal
// (ast.FuncLit). It's the caller's responsibility to find these and desugar
// them independently (if desired).
//
// At this time, desugaring is performed after packages have been loaded,
// parsed and type-checked, which means that any AST transformations below
// that introduce temporary variables must also update the associated
// types.Info. If this gets unruly in the future, desugaring should be
// performed after parsing AST's but before type checking so that this is
// done automatically by the type checker.
func desugar(stmt ast.Stmt, info *types.Info) ast.Stmt {
	d := desugarer{info: info}
	return d.desugar(stmt)
}

type desugarer struct {
	info   *types.Info
	vars   int
	labels int
}

func (d *desugarer) desugar(stmt ast.Stmt) ast.Stmt {
	switch s := stmt.(type) {
	case nil:
		return nil

	case *ast.BlockStmt:
		return &ast.BlockStmt{List: d.desugarList(s.List)}

	case *ast.IfStmt:
		// Rewrite `if init; cond {}` => `{ init; if cond {} }`
		s = &ast.IfStmt{
			Init: d.desugar(s.Init),
			Cond: s.Cond,
			Body: d.desugar(s.Body).(*ast.BlockStmt),
			Else: d.desugar(s.Else),
		}
		if init := s.Init; init != nil {
			s.Init = nil
			return &ast.BlockStmt{List: []ast.Stmt{init, s}}
		}
		return s

	case *ast.ForStmt:
		// Rewrite `for init; cond; post {}` => `{ init; for ; cond; post {} }`
		s = &ast.ForStmt{
			Init: d.desugar(s.Init),
			Cond: s.Cond,
			Body: d.desugar(s.Body).(*ast.BlockStmt),
			Post: d.desugar(s.Post),
		}
		if init := s.Init; init != nil {
			s.Init = nil
			return &ast.BlockStmt{List: []ast.Stmt{init, s}}
		}
		return s

	case *ast.RangeStmt:
		// Rewrite for range loops over arrays/slices.
		// - `for range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
		// - `for _ := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
		// - `for _, _ := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
		// - `for i := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ {} }`
		// - `for i, _ := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ {} }`
		// - `for i, v := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ { v := _x[i]; ... } }`
		// - `for _, v := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ { v := _x[_i]; ... } }`
		// Then, desugar loops further (see ast.ForStmt case above).
		var idx *ast.Ident
		if hasIdx := s.Key != nil && !isUnderscore(s.Key); !hasIdx {
			idx = d.newVar(types.Typ[types.Int])
		} else {
			idx = s.Key.(*ast.Ident)
		}
		x := d.newVar(d.info.TypeOf(s.X))
		if hasVal := s.Value != nil && !isUnderscore(s.Value); hasVal {
			s.Body.List = append([]ast.Stmt{
				&ast.AssignStmt{Lhs: []ast.Expr{s.Value}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.IndexExpr{X: x, Index: idx}}},
			}, s.Body.List...)
		}
		return d.desugar(&ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{Lhs: []ast.Expr{x}, Tok: token.DEFINE, Rhs: []ast.Expr{s.X}},
				&ast.ForStmt{
					Init: &ast.AssignStmt{Lhs: []ast.Expr{idx}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
					Post: &ast.IncDecStmt{X: idx, Tok: token.INC},
					Cond: &ast.BinaryExpr{X: idx, Op: token.LSS, Y: &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{x}}},
					Body: s.Body,
				},
			},
		})

	case *ast.SwitchStmt:
		// Rewrite `switch init; tag {}` to `init; switch tag {}`
		s = &ast.SwitchStmt{
			Init: d.desugar(s.Init),
			Tag:  s.Tag,
			Body: d.desugar(s.Body).(*ast.BlockStmt),
		}
		if init := s.Init; init != nil {
			s.Init = nil
			return &ast.BlockStmt{List: []ast.Stmt{init, s}}
		}
		return s

	case *ast.TypeSwitchStmt:
		// Rewrite `switch init; assign {}` to `init; switch assign {}`
		s = &ast.TypeSwitchStmt{
			Init:   d.desugar(s.Init),
			Assign: d.desugar(s.Assign),
			Body:   d.desugar(s.Body).(*ast.BlockStmt),
		}
		if init := s.Init; init != nil {
			s.Init = nil
			return &ast.BlockStmt{List: []ast.Stmt{init, s}}
		}
		return s

	case *ast.CaseClause:
		return &ast.CaseClause{
			List: s.List,
			Body: d.desugarList(s.Body),
		}

	case *ast.LabeledStmt:
		panic("not implemented")

	case *ast.SelectStmt, *ast.CommClause:
		panic("not implemented")

	case *ast.AssignStmt, *ast.BranchStmt, *ast.DeclStmt, *ast.DeferStmt, *ast.EmptyStmt,
		*ast.ExprStmt, *ast.GoStmt, *ast.IncDecStmt, *ast.ReturnStmt, *ast.SendStmt:
		return stmt
	}
	panic(fmt.Sprintf("unsupported ast.Stmt: %T", stmt))
}

func (d *desugarer) desugarList(stmts []ast.Stmt) []ast.Stmt {
	desugared := make([]ast.Stmt, len(stmts))
	for i, s := range stmts {
		desugared[i] = d.desugar(s)
	}
	return desugared
}

func (d *desugarer) newVar(t types.Type) *ast.Ident {
	v := ast.NewIdent("_v" + strconv.Itoa(d.vars))
	d.vars++
	d.info.Defs[v] = types.NewVar(0, nil, v.Name, t)
	return v
}

func (d *desugarer) newLabel() *ast.Ident {
	l := ast.NewIdent("_l" + strconv.Itoa(d.labels))
	d.labels++
	return l
}

func isUnderscore(e ast.Expr) bool {
	i, ok := e.(*ast.Ident)
	return ok && i.Name == "_"
}
