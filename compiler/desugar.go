package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
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
func desugar(p *packages.Package, stmt ast.Stmt, mayYield map[ast.Node]struct{}) ast.Stmt {
	d := desugarer{pkg: p, info: p.TypesInfo, nodesThatMayYield: mayYield}
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
	pkg               *packages.Package
	info              *types.Info
	vars              int
	labels            int
	nodesThatMayYield map[ast.Node]struct{}
	unusedLabels      map[*ast.Ident]struct{}
	userLabels        map[types.Object]*ast.Ident
}

func (d *desugarer) desugar(stmt ast.Stmt, breakTo, continueTo, userLabel *ast.Ident) ast.Stmt {
	if !d.mayYield(stmt) {
		return stmt
	}

	switch s := stmt.(type) {
	case nil:

	// These statements are desugared in flatMap(), since decomposing
	// expressions within them may require additional temporary variables
	// and thus additional assignment statements.
	case *ast.AssignStmt:
	case *ast.DeclStmt:
	case *ast.ExprStmt:
	case *ast.SendStmt:
	case *ast.ReturnStmt:
	case *ast.IncDecStmt:

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
			List: s.List, // desugared as part of the ast.SwitchStmt case
			Body: d.desugarList(s.Body, breakTo, continueTo),
		}

	case *ast.CommClause:
		stmt = &ast.CommClause{
			Comm: s.Comm, // desugared as part of the ast.SelectStmt case
			Body: d.desugarList(s.Body, breakTo, continueTo),
		}

	case *ast.DeferStmt:
		panic("not implemented")

	case *ast.EmptyStmt:

	case *ast.ForStmt:
		// Rewrite for statements:
		// - `for init; cond; post { ... }` => `{ init; for ; cond; post { ... } }`
		// - `for ; cond; post { ... }` => `for ; ; post { if !cond { break } ... }
		forLabel := d.newLabel()
		if userLabel != nil {
			d.addUserLabel(userLabel, forLabel)
		}
		body := &ast.BlockStmt{List: s.Body.List}
		if d.mayYield(s.Body) {
			d.nodesThatMayYield[body] = struct{}{}
		}
		if d.mayYield(s.Cond) {
			cond := &ast.UnaryExpr{Op: token.NOT, X: s.Cond}
			branch := &ast.BranchStmt{Tok: token.BREAK}
			guard := &ast.IfStmt{
				Cond: cond,
				Body: &ast.BlockStmt{List: []ast.Stmt{branch}},
			}
			d.nodesThatMayYield[cond] = struct{}{}
			d.nodesThatMayYield[branch] = struct{}{}
			d.nodesThatMayYield[guard] = struct{}{}
			d.nodesThatMayYield[guard.Body] = struct{}{}
			d.nodesThatMayYield[body] = struct{}{}
			body.List = append([]ast.Stmt{guard}, body.List...)
			s.Cond = nil
		}
		stmt = &ast.LabeledStmt{
			Label: forLabel,
			Stmt: &ast.ForStmt{
				Cond: s.Cond,
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
		if s.Init != nil {
			prologue := d.desugarList([]ast.Stmt{s.Init}, nil, nil)
			stmt = &ast.BlockStmt{List: append(prologue, stmt)}
		}

	case *ast.GoStmt:
		panic("not implemented")

	case *ast.IfStmt:
		// Rewrite `if init; cond { ... }` => `{ init; _cond := cond; if _cond { ... } }`
		var prologue []ast.Stmt
		if s.Init != nil {
			prologue = []ast.Stmt{s.Init}
		}
		if d.mayYield(s.Cond) {
			cond := d.newVar(types.Typ[types.Bool])
			assign := &ast.AssignStmt{
				Lhs: []ast.Expr{cond},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{s.Cond},
			}
			d.nodesThatMayYield[assign] = struct{}{}
			prologue = append(prologue, assign)
			s.Cond = cond
		}
		prologue = d.desugarList(prologue, nil, nil)
		stmt = &ast.BlockStmt{
			List: append(prologue, &ast.IfStmt{
				Cond: s.Cond,
				Body: d.desugar(s.Body, breakTo, continueTo, nil).(*ast.BlockStmt),
				Else: d.desugar(s.Else, breakTo, continueTo, nil),
			}),
		}

	case *ast.LabeledStmt:
		// Remove the user's label, but notify the next step so that generated
		// labels can be mapped.
		stmt = d.desugar(s.Stmt, breakTo, continueTo, s.Label)

	case *ast.RangeStmt:
		x := d.newVar(d.info.TypeOf(s.X))
		init := &ast.AssignStmt{Lhs: []ast.Expr{x}, Tok: token.DEFINE, Rhs: []ast.Expr{s.X}}
		if d.mayYield(s.X) {
			d.nodesThatMayYield[init] = struct{}{}
		}
		prologue := d.desugarList([]ast.Stmt{init}, nil, nil)

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
			forStmt := &ast.ForStmt{
				Init: &ast.AssignStmt{Lhs: []ast.Expr{i}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
				Post: &ast.IncDecStmt{X: i, Tok: token.INC},
				Cond: &ast.BinaryExpr{X: i, Op: token.LSS, Y: &ast.CallExpr{Fun: d.builtin("len"), Args: []ast.Expr{x}}},
				Body: s.Body,
			}
			if d.mayYield(s.Body) {
				d.nodesThatMayYield[forStmt] = struct{}{}
			}
			stmt = &ast.BlockStmt{
				List: append(prologue, d.desugar(forStmt, breakTo, continueTo, userLabel)),
			}

		case *types.Map:
			// Handle the simple case first:
			if (s.Key == nil || isUnderscore(s.Key)) && (s.Value == nil || isUnderscore(s.Value)) {
				// Rewrite `for range m {}` => `{ _x := m; for _i := 0; _i < len(_x); _i++ {} }`
				i := d.newVar(types.Typ[types.Int])
				forStmt := &ast.ForStmt{
					Init: &ast.AssignStmt{Lhs: []ast.Expr{i}, Tok: token.DEFINE, Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}}},
					Post: &ast.IncDecStmt{X: i, Tok: token.INC},
					Cond: &ast.BinaryExpr{X: i, Op: token.LSS, Y: &ast.CallExpr{Fun: d.builtin("len"), Args: []ast.Expr{x}}},
					Body: s.Body,
				}
				if d.mayYield(s.Body) {
					d.nodesThatMayYield[forStmt] = struct{}{}
				}
				stmt = &ast.BlockStmt{
					List: append(prologue, d.desugar(forStmt, breakTo, continueTo, userLabel)),
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
								Fun: d.builtin("make"),
								Args: []ast.Expr{
									typeExpr(d.pkg, keySliceType),
									&ast.BasicLit{Kind: token.INT, Value: "0"},
									&ast.CallExpr{Fun: d.builtin("len"), Args: []ast.Expr{x}},
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
											&ast.CallExpr{Fun: d.builtin("append"), Args: []ast.Expr{keys, k}},
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
				guard := &ast.IfStmt{
					Init: &ast.AssignStmt{
						Lhs: []ast.Expr{mapValue, ok},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.IndexExpr{X: x, Index: mapKey}},
					},
					Cond: ok,
					Body: s.Body,
				}
				rangeLoop := &ast.RangeStmt{
					Value: mapKey,
					Tok:   token.DEFINE,
					X:     keys,
					Body: &ast.BlockStmt{
						List: []ast.Stmt{guard},
					},
				}
				if d.mayYield(s.Body) {
					d.nodesThatMayYield[guard] = struct{}{}
					d.nodesThatMayYield[rangeLoop.Body] = struct{}{}
					d.nodesThatMayYield[rangeLoop] = struct{}{}
				}
				iterKeys := d.desugar(rangeLoop, breakTo, continueTo, userLabel)

				stmt = &ast.BlockStmt{List: append(prologue, collectKeys, iterKeys)}
			}
		default:
			panic(fmt.Sprintf("not implemented: for range over %T", s.X))
		}

	case *ast.SelectStmt:
		if s.Body.List == nil {
			return &ast.SelectStmt{Body: &ast.BlockStmt{}}
		}

		// Rewrite select statements into a select+switch statement. The
		// select cases exist only to record the selection; the select
		// case bodies are moved into the switch statement over that
		// selection. This allows coroutines to jump back to the right
		// case when resuming.
		selection := d.newVar(types.Typ[types.Int])
		prologue := []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{selection},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
			},
		}
		rawSelect := &ast.SelectStmt{Body: &ast.BlockStmt{List: make([]ast.Stmt, len(s.Body.List))}}
		switchBody := &ast.BlockStmt{List: make([]ast.Stmt, len(s.Body.List))}
		switchStmt := &ast.SwitchStmt{Tag: selection, Body: switchBody}

		for i, c := range s.Body.List {
			cc := c.(*ast.CommClause)
			caseComm := cc.Comm
			id := &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i + 1)}
			switchBody.List[i] = &ast.CaseClause{
				List: []ast.Expr{id},
				Body: cc.Body,
			}
			for _, n := range cc.Body {
				if d.mayYield(n) {
					d.nodesThatMayYield[switchBody.List[i]] = struct{}{}
					d.nodesThatMayYield[switchBody] = struct{}{}
					d.nodesThatMayYield[switchStmt] = struct{}{}
					break
				}
			}
			switch m := caseComm.(type) {
			case nil:
			case *ast.SendStmt:
				tmpChan := d.newVar(d.info.TypeOf(m.Chan))
				tmpValue := d.newVar(d.info.TypeOf(m.Value))
				assignChan := &ast.AssignStmt{Lhs: []ast.Expr{tmpChan}, Tok: token.DEFINE, Rhs: []ast.Expr{m.Chan}}
				assignValue := &ast.AssignStmt{Lhs: []ast.Expr{tmpValue}, Tok: token.DEFINE, Rhs: []ast.Expr{m.Value}}
				if d.mayYield(m.Chan) {
					d.nodesThatMayYield[assignChan] = struct{}{}
				}
				if d.mayYield(assignValue) {
					d.nodesThatMayYield[assignValue] = struct{}{}
				}
				prologue = append(prologue, assignChan, assignValue)
				m.Chan = tmpChan
				m.Value = tmpValue
			case *ast.ExprStmt:
				recv := m.X.(*ast.UnaryExpr)
				if recv.Op != token.ARROW {
					panic("unexpected select case")
				}
				tmpRecv := d.newVar(d.info.TypeOf(recv.X))
				assignRecv := &ast.AssignStmt{Lhs: []ast.Expr{tmpRecv}, Tok: token.DEFINE, Rhs: []ast.Expr{recv.X}}
				if d.mayYield(assignRecv) {
					d.nodesThatMayYield[assignRecv] = struct{}{}
				}
				prologue = append(prologue, assignRecv)
				recv.X = tmpRecv
			case *ast.AssignStmt:
				if len(m.Rhs) != 1 {
					panic("unexpected select case")
				}
				recv := m.Rhs[0].(*ast.UnaryExpr)
				if recv.Op != token.ARROW {
					panic("unexpected select case")
				}
				tmpRecv := d.newVar(d.info.TypeOf(recv.X))
				assignRecv := &ast.AssignStmt{Lhs: []ast.Expr{tmpRecv}, Tok: token.DEFINE, Rhs: []ast.Expr{recv.X}}
				if d.mayYield(assignRecv) {
					d.nodesThatMayYield[assignRecv] = struct{}{}
				}
				prologue = append(prologue, assignRecv)
				recv.X = tmpRecv
				caseBodyAssigns := make([]ast.Stmt, len(m.Lhs))
				for j, lhs := range m.Lhs {
					lhsType := d.info.TypeOf(lhs)
					tmpLhs := d.newVar(lhsType)
					prologue = append(prologue,
						&ast.DeclStmt{Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{tmpLhs},
									Type:  typeExpr(d.pkg, lhsType),
								},
							},
						}})

					caseBodyAssigns[j] = &ast.AssignStmt{Lhs: []ast.Expr{lhs}, Tok: m.Tok, Rhs: []ast.Expr{tmpLhs}}
					if d.mayYield(lhs) {
						d.nodesThatMayYield[caseBodyAssigns[j]] = struct{}{}
						d.nodesThatMayYield[switchBody.List[i]] = struct{}{}
						d.nodesThatMayYield[switchBody] = struct{}{}
						d.nodesThatMayYield[switchStmt] = struct{}{}
					}
					m.Lhs[j] = tmpLhs
				}
				switchBodyCase := switchBody.List[i].(*ast.CaseClause)
				switchBodyCase.Body = append(caseBodyAssigns, switchBodyCase.Body...)
				m.Tok = token.ASSIGN
			default:
				panic(fmt.Sprintf("unexpected select case %T", m))
			}

			rawSelect.Body.List[i] = &ast.CommClause{
				Comm: caseComm,
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{selection},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{id},
					},
				},
			}
		}
		prologue = d.desugarList(prologue, nil, nil)
		stmt = &ast.BlockStmt{
			List: append(prologue,
				rawSelect,
				d.desugar(switchStmt, breakTo, continueTo, userLabel),
			),
		}

	case *ast.SwitchStmt:
		// Rewrite switch statements:
		// - `switch init; tag { ... }` => `{ init; _tag := tag; switch _tag { ... }`
		switchLabel := d.newLabel()
		if userLabel != nil {
			d.addUserLabel(userLabel, switchLabel)
		}
		var prologue []ast.Stmt
		if s.Init != nil {
			prologue = []ast.Stmt{s.Init}
		}
		var tag ast.Expr
		if s.Tag != nil {
			tag = d.newVar(d.info.TypeOf(s.Tag))
			assign := &ast.AssignStmt{
				Lhs: []ast.Expr{tag},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{s.Tag},
			}
			if d.mayYield(s.Tag) {
				d.nodesThatMayYield[assign] = struct{}{}
			}
			prologue = append(prologue, assign)
		}
		var defaultCaseBody ast.Stmt
		var head ast.Stmt
		var tail *ast.IfStmt
		for _, caseStmt := range s.Body.List {
			c := caseStmt.(*ast.CaseClause)
			bodyMayYield := false
			for _, n := range c.Body {
				if d.mayYield(n) {
					bodyMayYield = true
					break
				}
			}
			if len(c.List) == 0 {
				defaultCaseBody = &ast.BlockStmt{List: c.Body}
				if bodyMayYield {
					d.nodesThatMayYield[defaultCaseBody] = struct{}{}
				}
				continue
			}
			list := make([]ast.Expr, len(c.List))
			for i := range list {
				value := c.List[i]
				if tag != nil {
					list[i] = &ast.BinaryExpr{X: tag, Op: token.EQL, Y: value}
					if d.mayYield(value) {
						d.nodesThatMayYield[list[i]] = struct{}{}
					}
				} else {
					list[i] = value
				}
			}
			tmp := d.newVar(types.Typ[types.Bool])
			orExpr := list[0]
			list = list[1:]
			for len(list) > 0 {
				// TODO: balance the tree
				x, y := orExpr, list[0]
				orExpr = &ast.BinaryExpr{X: x, Op: token.OR, Y: y}
				if d.mayYield(x) || d.mayYield(y) {
					d.nodesThatMayYield[orExpr] = struct{}{}
				}
				list = list[1:]
			}
			ifStmt := &ast.IfStmt{
				Init: &ast.AssignStmt{Lhs: []ast.Expr{tmp}, Tok: token.DEFINE, Rhs: []ast.Expr{orExpr}},
				Cond: tmp,
				Body: &ast.BlockStmt{List: c.Body},
			}
			if d.mayYield(orExpr) {
				d.nodesThatMayYield[ifStmt.Init] = struct{}{}
				d.nodesThatMayYield[ifStmt] = struct{}{}
			}
			if bodyMayYield {
				d.nodesThatMayYield[ifStmt.Body] = struct{}{}
				d.nodesThatMayYield[ifStmt] = struct{}{}
			}
			if head == nil {
				head = ifStmt
				tail = ifStmt
			} else {
				tail.Else = ifStmt
				tail = ifStmt
			}
		}
		if defaultCaseBody != nil {
			if head == nil {
				head = defaultCaseBody
			} else {
				tail.Else = defaultCaseBody
			}
		}
		if head == nil {
			head = &ast.EmptyStmt{}
		} else {
			s.Tag = nil
		}

		prologue = d.desugarList(prologue, nil, nil)

		stmt = &ast.LabeledStmt{
			Label: switchLabel,
			Stmt: &ast.SwitchStmt{
				Tag: s.Tag,
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.CaseClause{
							Body: []ast.Stmt{
								d.desugar(head, switchLabel, continueTo, nil),
							},
						},
					},
				},
			},
		}
		if len(prologue) > 0 {
			stmt = &ast.BlockStmt{List: append(prologue, stmt)}
		}

	case *ast.TypeSwitchStmt:
		// Rewrite type switch statements:
		// - `switch init; x.(type) { ... }` to `{ init; _x := x; switch _x.(type) { ... } }`
		// - `switch init; x := y.(type) { ... }` to `{ init; _t := y; switch x := _y.(type) { ... } }`
		switchLabel := d.newLabel()
		if userLabel != nil {
			d.addUserLabel(userLabel, switchLabel)
		}
		var prologue []ast.Stmt
		if s.Init != nil {
			prologue = []ast.Stmt{s.Init}
		}

		// https://go.dev/ref/spec#TypeSwitchStmt
		var t *ast.TypeAssertExpr
		switch a := s.Assign.(type) {
		case *ast.ExprStmt:
			t = a.X.(*ast.TypeAssertExpr)
		case *ast.AssignStmt:
			t = a.Rhs[0].(*ast.TypeAssertExpr)
		}
		if d.mayYield(t.X) {
			tmp := d.newVar(d.info.TypeOf(t.X))
			prologue = append(prologue, &ast.AssignStmt{
				Lhs: []ast.Expr{tmp},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{t.X},
			})
			t.X = tmp
		}

		prologue = d.desugarList(prologue, nil, nil)
		stmt = &ast.BlockStmt{
			List: append(prologue, &ast.LabeledStmt{
				Label: switchLabel,
				Stmt: &ast.TypeSwitchStmt{
					Assign: s.Assign,
					Body:   d.desugar(s.Body, switchLabel, continueTo, nil).(*ast.BlockStmt),
				},
			}),
		}

	default:
		panic(fmt.Sprintf("unsupported ast.Stmt: %T", stmt))
	}
	return stmt
}

func (d *desugarer) desugarList(stmts []ast.Stmt, breakTo, continueTo *ast.Ident) []ast.Stmt {
	desugared := make([]ast.Stmt, 0, len(stmts))
	for _, s := range stmts {
		gen := d.flatMap(s)
		for _, gs := range gen {
			desugared = append(desugared, d.desugar(gs, breakTo, continueTo, nil))
		}
	}
	return desugared
}

func (d *desugarer) flatMap(stmt ast.Stmt) (result []ast.Stmt) {
	var prereqs []ast.Stmt
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		var flags exprFlags
		if s.Tok == token.DEFINE {
			// LHS is just ast.Ident in this case; no need to decompose.
			if len(s.Rhs) > 1 {
				flags |= multiExprStmt
			}
		} else {
			flags |= multiExprStmt
			for j, expr := range s.Lhs {
				s.Lhs[j], prereqs = d.decomposeExpression(expr, flags)
				result = append(result, prereqs...)
			}
		}
		for j, expr := range s.Rhs {
			s.Rhs[j], prereqs = d.decomposeExpression(expr, flags)
			result = append(result, prereqs...)
		}
	case *ast.DeclStmt:
		g := s.Decl.(*ast.GenDecl)
		if g.Tok == token.VAR {
			for _, spec := range g.Specs {
				v := spec.(*ast.ValueSpec)
				var flags exprFlags
				if len(v.Values) > 1 {
					flags |= multiExprStmt
				}
				for j, expr := range v.Values {
					v.Values[j], prereqs = d.decomposeExpression(expr, flags)
					result = append(result, prereqs...)
				}
			}
		}
	case *ast.ExprStmt:
		s.X, prereqs = d.decomposeExpression(s.X, exprFlags(0))
		result = append(result, prereqs...)
	case *ast.SendStmt:
		s.Chan, prereqs = d.decomposeExpression(s.Chan, multiExprStmt)
		result = append(result, prereqs...)
		s.Value, prereqs = d.decomposeExpression(s.Value, multiExprStmt)
		result = append(result, prereqs...)
	case *ast.ReturnStmt:
		var flags exprFlags
		if len(s.Results) > 1 {
			flags |= multiExprStmt
		}
		for j, expr := range s.Results {
			s.Results[j], prereqs = d.decomposeExpression(expr, flags)
			result = append(result, prereqs...)
		}
	case *ast.IncDecStmt:
		s.X, prereqs = d.decomposeExpression(s.X, exprFlags(0))
		result = append(result, prereqs...)
	}
	result = append(result, stmt)
	return
}

type exprFlags int

const (
	// multiExprStmt is set if the expression is part of a statement
	// that has more than one nested expression of type ast.Expr.
	multiExprStmt exprFlags = 1 << iota
)

func (d *desugarer) mayYield(n ast.Node) bool {
	switch n.(type) {
	case nil:
		return false
	case *ast.BasicLit, *ast.FuncLit, *ast.Ident:
		return false
	case *ast.ArrayType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType, *ast.MapType, *ast.StructType:
		return false
	}
	_, ok := d.nodesThatMayYield[n]
	return ok
}

func (d *desugarer) decomposeExpression(expr ast.Expr, flags exprFlags) (ast.Expr, []ast.Stmt) {
	if !d.mayYield(expr) {
		return expr, nil
	}

	queue := []ast.Expr{expr}
	var tmps []*ast.Ident

	decompose := func(e ast.Expr) ast.Expr {
		if !d.mayYield(e) {
			return e
		}
		tmp := d.newVar(d.info.TypeOf(e))
		tmps = append(tmps, tmp)
		queue = append(queue, e)
		return tmp
	}

	for i := 0; i < len(queue); i++ {
		switch e := queue[i].(type) {
		case *ast.BadExpr:
			panic("bad expr")

		case *ast.BinaryExpr:
			e.X = decompose(e.X)
			e.Y = decompose(e.Y)

		case *ast.CallExpr:
			if i == 0 && (flags&multiExprStmt) != 0 {
				// Need to hoist the CallExpr out into a temporary variable in
				// this case, so that the relative order of calls (and their
				// prerequisites) is preserved.
				switch d.info.TypeOf(e).(type) {
				case *types.Tuple:
					// TODO: can't hoist like this when it's a function
					//  that returns multiple values
				default:
					queue[i] = decompose(e)
					continue
				}
			}
			e.Fun = decompose(e.Fun)
			for i, arg := range e.Args {
				e.Args[i] = decompose(arg)
			}

		case *ast.CompositeLit:
			for i, elt := range e.Elts {
				e.Elts[i] = decompose(elt)
			}
			// skip e.Type (type expression)

		case *ast.Ellipsis:
			e.Elt = decompose(e.Elt)

		case *ast.IndexExpr:
			e.X = decompose(e.X)
			e.Index = decompose(e.Index)

		case *ast.IndexListExpr:
			e.X = decompose(e.X)
			// skip e.Indices (type expressions)

		case *ast.KeyValueExpr:
			e.Key = decompose(e.Key)
			e.Value = decompose(e.Value)

		case *ast.ParenExpr:
			e.X = decompose(e.X)

		case *ast.SelectorExpr:
			e.X = decompose(e.X)

		case *ast.SliceExpr:
			e.X = decompose(e.X)
			e.Low = decompose(e.Low)
			e.Max = decompose(e.Max)
			e.High = decompose(e.High)

		case *ast.StarExpr:
			e.X = decompose(e.X)

		case *ast.TypeAssertExpr:
			e.X = decompose(e.X)
			// skip e.Type (type expression)

		case *ast.UnaryExpr:
			e.X = decompose(e.X)

		default:
			panic(fmt.Sprintf("unsupported ast.Expr: %T", queue[i]))
		}
	}
	prereqs := make([]ast.Stmt, len(tmps))
	for i := range tmps {
		prereqs[i] = &ast.AssignStmt{
			Lhs: []ast.Expr{tmps[i]},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{queue[i+1]},
		}
	}
	reverse(prereqs)
	return queue[0], prereqs
}

func reverse(stmts []ast.Stmt) {
	i := 0
	j := len(stmts) - 1
	for i < j {
		stmts[i], stmts[j] = stmts[j], stmts[i]
		i++
		j--
	}
}

func (d *desugarer) builtin(name string) *ast.Ident {
	ident := ast.NewIdent(name)
	d.info.Uses[ident] = types.Universe.Lookup(name)
	return ident
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

// findCalls marks nodes in a tree that are an *ast.CallExpr, or lead to
// an *ast.CallExpr.
func findCalls(tree ast.Node, info *types.Info) map[ast.Node]struct{} {
	mayYield := map[ast.Node]struct{}{}
	var stack []ast.Node
	ast.Inspect(tree, func(node ast.Node) bool {
		if node != nil {
			stack = append(stack, node)

			if c, ok := node.(*ast.CallExpr); ok {
				// Exclude some call expressions.
				switch fn := c.Fun.(type) {
				case *ast.Ident:
					if obj := info.ObjectOf(fn); obj != nil {
						if obj == types.Universe.Lookup(fn.Name) {
							return true // skip builtin function calls
						} else if _, ok := obj.(*types.TypeName); ok {
							return true // skip type casts
						}
					}
				}

				// Mark this node, and all nodes that lead to it.
			addNodes:
				for i := len(stack) - 1; i >= 0; i-- {
					n := stack[i]
					switch n.(type) {
					case *ast.FuncDecl, *ast.FuncLit:
						break addNodes
					}
					if _, ok := mayYield[n]; ok {
						break
					}
					mayYield[n] = struct{}{}
				}
			}
		} else {
			stack = stack[:len(stack)-1]
		}
		return true
	})

	return mayYield
}

// markBranchStmt marks nodes in a tree that are of type
// ast.BranchStmt and that have an ancestor in the specified set.
// When marking nodes, all ancestors are also marked.
func markBranchStmt(tree ast.Node, set map[ast.Node]struct{}) {
	var stack []ast.Node
	ast.Inspect(tree, func(node ast.Node) bool {
		if node != nil {
			stack = append(stack, node)
			if _, ok := node.(*ast.BranchStmt); ok {
				closestAncestor := -1
			findAncestor:
				for i := len(stack) - 1; i >= 0; i-- {
					n := stack[i]
					switch n.(type) {
					case *ast.FuncDecl, *ast.FuncLit:
						break findAncestor
					}
					if _, ok := set[n]; ok {
						closestAncestor = i
						break
					}
				}
				if closestAncestor >= 0 {
					for i := closestAncestor; i < len(stack); i++ {
						set[stack[i]] = struct{}{}
					}
				}
			}
		} else {
			stack = stack[:len(stack)-1]
		}
		return true
	})
}
