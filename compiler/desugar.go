package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
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
// Nondeterministic control flow and iteration (e.g. select, for..range
// over maps) is split into two parts so that yield points within can resume
// from the same place.
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
	stmt = d.desugar(stmt, nil, nil, nil)

	// Unused labels cause a compile error (label X defined and not used)
	// so we need a second pass over the tree to delete unused labels.
	astutil.Apply(stmt, func(cursor *astutil.Cursor) bool {
		if ls, ok := cursor.Node().(*ast.LabeledStmt); ok && d.isUnusedLabel(ls.Label) {
			cursor.Replace(ls.Stmt)
		}
		return true
	}, nil)

	return stmt
}

type desugarer struct {
	info         *types.Info
	vars         int
	labels       int
	unusedLabels map[*ast.Ident]struct{}
	userLabels   map[types.Object]*ast.Ident
}

func (d *desugarer) desugar(stmt ast.Stmt, breakTo, continueTo, userLabel *ast.Ident) ast.Stmt {
	switch s := stmt.(type) {
	case nil:

	case *ast.AssignStmt:
		// TODO: desugar expressions

	case *ast.BadStmt:
		panic("bad stmt")

	case *ast.BlockStmt:
		stmt = &ast.BlockStmt{List: d.desugarList(s.List, breakTo, continueTo)}

	case *ast.BranchStmt:
		if s.Label != nil {
			label := d.getUserLabel(s.Label)
			if label == nil {
				panic(fmt.Sprintf("label not found: %s", s.Label))
			}
			d.useLabel(label)
			stmt = &ast.BranchStmt{Tok: s.Tok, Label: label}
		} else {
			switch s.Tok {
			case token.BREAK:
				d.useLabel(breakTo)
				stmt = &ast.BranchStmt{Tok: token.BREAK, Label: breakTo}
			case token.CONTINUE:
				d.useLabel(continueTo)
				stmt = &ast.BranchStmt{Tok: token.CONTINUE, Label: continueTo}
			default: // FALLTHROUGH / GOTO
				panic("not implemented")
			}
		}

	case *ast.CaseClause:
		stmt = &ast.CaseClause{
			List: s.List, // desugared in SwitchStmt case
			Body: d.desugarList(s.Body, breakTo, continueTo),
		}

	case *ast.CommClause:
		stmt = &ast.CommClause{
			Comm: d.desugar(s.Comm, nil, nil, nil), // partially desugared in SelectStmt case
			Body: d.desugarList(s.Body, breakTo, continueTo),
		}

	case *ast.DeclStmt:
		// TODO: desugar expressions on RHS of var decls

	case *ast.DeferStmt:
		panic("not implemented")

	case *ast.EmptyStmt:

	case *ast.ExprStmt:
		// TODO: desugar expressions

	case *ast.ForStmt:
		// Rewrite for statements:
		// - `for init; cond; post { ... }` => `{ init; for ; cond; post { ... } }`
		// - `for ; cond; post { ... }` => `for ; ; post { if !cond { break } ... }
		init := d.desugar(s.Init, nil, nil, nil)
		forLabel := d.newLabel()
		if userLabel != nil {
			d.addUserLabel(userLabel, forLabel)
		}
		body := &ast.BlockStmt{List: s.Body.List}
		if s.Cond != nil {
			body.List = append([]ast.Stmt{&ast.IfStmt{
				Cond: &ast.UnaryExpr{Op: token.NOT, X: s.Cond},
				Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}},
			}}, body.List...)
		}
		stmt = &ast.LabeledStmt{
			Label: forLabel,
			Stmt: &ast.ForStmt{
				Body: d.desugar(body, forLabel, forLabel, nil).(*ast.BlockStmt),
				// The post iteration statement is currently preserved for a
				// later pass.
				// TODO: find a way to move the statement into the loop body
				//  so that it can be desugared further, but do so in a way
				//  that doesn't break continue and deeply nested cases of
				//  continue [label]. Using goto doesn't work!
				Post: d.desugar(s.Post, nil, nil, nil),
			},
		}
		if init != nil {
			stmt = &ast.BlockStmt{List: []ast.Stmt{init, stmt}}
		}

	case *ast.GoStmt:
		panic("not implemented")

	case *ast.IfStmt:
		// Rewrite `if init; cond { ... }` => `{ init; _cond := cond; if _cond { ... } }`
		init := d.desugar(s.Init, nil, nil, nil)
		condvar := d.newVar(types.Typ[types.Bool])
		cond := &ast.AssignStmt{
			Lhs: []ast.Expr{condvar},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{s.Cond},
		}
		var prologue []ast.Stmt
		if init != nil {
			prologue = []ast.Stmt{init, cond}
		} else {
			prologue = []ast.Stmt{cond}
		}
		prologue = d.desugarList(prologue, nil, nil)
		stmt = &ast.BlockStmt{
			List: append(prologue, &ast.IfStmt{
				Cond: condvar,
				Body: d.desugar(s.Body, breakTo, continueTo, nil).(*ast.BlockStmt),
				Else: d.desugar(s.Else, breakTo, continueTo, nil),
			}),
		}

	case *ast.IncDecStmt:
		// TODO: desugar expressions

	case *ast.LabeledStmt:
		// Remove the user's label, but notify the next step so that generated
		// labels can be mapped.
		stmt = d.desugar(s.Stmt, breakTo, continueTo, s.Label)

	case *ast.RangeStmt:
		x := d.newVar(d.info.TypeOf(s.X))
		init := &ast.AssignStmt{Lhs: []ast.Expr{x}, Tok: token.DEFINE, Rhs: []ast.Expr{s.X}}

		switch rangeElemType := d.info.TypeOf(s.X).(type) {
		case *types.Array, *types.Slice:
			// Rewrite for range loops over arrays/slices:
			// - `for range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
			// - `for _ := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
			// - `for _, _ := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ {} }`
			// - `for i := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ {} }`
			// - `for i, _ := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ {} }`
			// - `for i, v := range x {}` => `{ _x := x; for i := 0; i < len(_x); i++ { v := _x[i]; ... } }`
			// - `for _, v := range x {}` => `{ _x := x; for _i := 0; _i < len(_x); _i++ { v := _x[_i]; ... } }`
			// Then, desugar loops further (see ast.ForStmt case above).
			var i *ast.Ident
			if s.Key == nil || isUnderscore(s.Key) {
				i = d.newVar(types.Typ[types.Int])
			} else {
				i = s.Key.(*ast.Ident)
			}
			if s.Value != nil && !isUnderscore(s.Value) {
				s.Body.List = append([]ast.Stmt{
					&ast.AssignStmt{Lhs: []ast.Expr{s.Value}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.IndexExpr{X: x, Index: i}}},
				}, s.Body.List...)
			}
			stmt = &ast.BlockStmt{
				List: []ast.Stmt{
					init,
					d.desugar(&ast.ForStmt{
						Init: &ast.AssignStmt{Lhs: []ast.Expr{i}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
						Post: &ast.IncDecStmt{X: i, Tok: token.INC},
						Cond: &ast.BinaryExpr{X: i, Op: token.LSS, Y: &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{x}}},
						Body: s.Body,
					}, breakTo, continueTo, userLabel),
				},
			}

		case *types.Map:
			// Handle the simple case first:
			if (s.Key == nil || isUnderscore(s.Key)) && (s.Value == nil || isUnderscore(s.Value)) {
				// Rewrite `for range m {}` => `{ _x := m; for _i := 0; _i < len(_x); _i++ {} }`
				i := d.newVar(types.Typ[types.Int])
				stmt = &ast.BlockStmt{
					List: []ast.Stmt{
						init,
						d.desugar(&ast.ForStmt{
							Init: &ast.AssignStmt{Lhs: []ast.Expr{i}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
							Post: &ast.IncDecStmt{X: i, Tok: token.INC},
							Cond: &ast.BinaryExpr{X: i, Op: token.LSS, Y: &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{x}}},
							Body: s.Body,
						}, breakTo, continueTo, userLabel),
					},
				}
			} else {
				// Since map iteration order is not deterministic, we split the
				// loop into two. The first loop collects keys, and the second
				// loop iterates over those keys.
				keyType := rangeElemType.Key()
				keySliceType := types.NewSlice(keyType)
				keys := d.newVar(keySliceType)

				k := d.newVar(types.Typ[types.Int])
				collectKeys := &ast.BlockStmt{
					List: []ast.Stmt{
						// _keys := make([]keyType, 0, len(_map))
						&ast.AssignStmt{Lhs: []ast.Expr{keys}, Tok: token.DEFINE, Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: ast.NewIdent("make"),
								Args: []ast.Expr{
									typeExpr(keySliceType),
									&ast.BasicLit{Kind: token.INT, Value: "0"},
									&ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{x}},
								},
							},
						}},
						// for k := range _map
						// Note that this loop isn't desugared!
						&ast.RangeStmt{
							Key: k,
							Tok: token.DEFINE,
							X:   x,
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									// _keys = append(_keys, k)
									&ast.AssignStmt{
										Lhs: []ast.Expr{keys},
										Tok: token.ASSIGN,
										Rhs: []ast.Expr{
											&ast.CallExpr{Fun: ast.NewIdent("append"), Args: []ast.Expr{keys, k}},
										},
									},
								},
							},
						},
					},
				}

				var mapKey *ast.Ident
				if s.Key == nil || isUnderscore(s.Key) {
					mapKey = d.newVar(keyType)
				} else {
					mapKey = s.Key.(*ast.Ident)
				}
				var mapValue *ast.Ident
				if s.Value != nil {
					mapValue = s.Value.(*ast.Ident)
				} else {
					mapValue = ast.NewIdent("_")
				}
				ok := d.newVar(types.Typ[types.Bool])
				iterKeys := d.desugar(&ast.RangeStmt{
					Value: mapKey,
					Tok:   token.DEFINE,
					X:     keys,
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.IfStmt{
								Init: &ast.AssignStmt{
									Lhs: []ast.Expr{mapValue, ok},
									Tok: token.DEFINE,
									Rhs: []ast.Expr{&ast.IndexExpr{X: x, Index: mapKey}},
								},
								Cond: ok,
								Body: s.Body,
							},
						},
					},
				}, breakTo, continueTo, userLabel)

				stmt = &ast.BlockStmt{List: []ast.Stmt{init, collectKeys, iterKeys}}
			}
		default:
			panic(fmt.Sprintf("not implemented: for range over %T", s.X))
		}

	case *ast.ReturnStmt:
		// TODO: desugar expressions

	case *ast.SelectStmt:
		// Rewrite select statements into a select+switch statement. The
		// select cases exist only to record the selection; the select
		// case bodies are moved into the switch statement over that
		// selection. This allows coroutines to jump back to the right
		// case when resuming.
		selection := d.newVar(types.Typ[types.Int])
		switchBody := &ast.BlockStmt{List: make([]ast.Stmt, len(s.Body.List))}
		for i, c := range s.Body.List {
			cc := c.(*ast.CommClause)
			// TODO: hoist cc.Comm out from select{} so expressions can be desugared
			id := &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i + 1)}
			switchBody.List[i] = &ast.CaseClause{
				List: []ast.Expr{id},
				Body: cc.Body,
			}
			cc.Body = []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{selection},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{id},
				},
			}
		}
		stmt = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{selection},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
				s,
				d.desugar(&ast.SwitchStmt{Tag: selection, Body: switchBody}, breakTo, continueTo, userLabel),
			},
		}

	case *ast.SendStmt:
		// TODO: desugar expressions

	case *ast.SwitchStmt:
		// Rewrite switch statements:
		// - `switch init; tag { case A: ... case B: ... }` => `{ init; if _tag := tag; _tag == A { ... } else if _tag == B { ... }`
		// - `switch { case A: ... case B: ... default: ... } => `if A { ... } else if B { ... } else { ... }`
		init := d.desugar(s.Init, nil, nil, nil)
		switchLabel := d.newLabel()
		if userLabel != nil {
			d.addUserLabel(userLabel, switchLabel)
		}
		// TODO: hoist each CaseClause.Cond out from SwitchStmt.Body so expressions can be desugared
		stmt = &ast.LabeledStmt{
			Label: switchLabel,
			Stmt: &ast.SwitchStmt{
				Tag:  s.Tag,
				Body: d.desugar(s.Body, switchLabel, continueTo, nil).(*ast.BlockStmt),
			},
		}
		if init != nil {
			stmt = &ast.BlockStmt{List: []ast.Stmt{init, stmt}}
		}

	case *ast.TypeSwitchStmt:
		// Rewrite type switch statements:
		// - `switch init; x.(type) { ... }` to `{ init; _x := x; switch _x.(type) { ... } }`
		// - `switch init; x := y.(type) { ... }` to `{ init; _t := y; switch x := _y.(type) { ... } }`
		init := d.desugar(s.Init, nil, nil, nil)
		switchLabel := d.newLabel()
		if userLabel != nil {
			d.addUserLabel(userLabel, switchLabel)
		}
		stmt = &ast.LabeledStmt{
			Label: switchLabel,
			Stmt: &ast.TypeSwitchStmt{
				Assign: d.desugar(s.Assign, nil, nil, nil),
				Body:   d.desugar(s.Body, switchLabel, continueTo, nil).(*ast.BlockStmt),
			},
		}
		if init != nil {
			stmt = &ast.BlockStmt{List: []ast.Stmt{init, stmt}}
		}

	default:
		panic(fmt.Sprintf("unsupported ast.Stmt: %T", stmt))
	}
	return stmt
}

func (d *desugarer) desugarList(stmts []ast.Stmt, breakTo, continueTo *ast.Ident) []ast.Stmt {
	desugared := make([]ast.Stmt, len(stmts))
	for i, s := range stmts {
		desugared[i] = d.desugar(s, breakTo, continueTo, nil)
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

	// Mark labels as unused initially.
	if d.unusedLabels == nil {
		d.unusedLabels = map[*ast.Ident]struct{}{}
	}
	d.unusedLabels[l] = struct{}{}

	return l
}

func (d *desugarer) addUserLabel(userLabel, replacement *ast.Ident) {
	if d.userLabels == nil {
		d.userLabels = map[types.Object]*ast.Ident{}
	}
	d.userLabels[d.info.ObjectOf(userLabel)] = replacement
}

func (d *desugarer) getUserLabel(userLabel *ast.Ident) *ast.Ident {
	return d.userLabels[d.info.ObjectOf(userLabel)]
}

func (d *desugarer) useLabel(label *ast.Ident) {
	delete(d.unusedLabels, label)
}

func (d *desugarer) isUnusedLabel(label *ast.Ident) bool {
	_, ok := d.unusedLabels[label]
	return ok
}

func isUnderscore(e ast.Expr) bool {
	i, ok := e.(*ast.Ident)
	return ok && i.Name == "_"
}
