package compiler

import (
	"go/ast"
	"go/token"
	"strconv"
)

// trackDispatchSpans assigns a non-zero monotonically increasing integer ID to each
// leaf statement in the tree using a post-order traversal, and then assigns
// a "span" to all statements in the tree which is equal to the half-open
// range of IDs seen in that subtree.
//
// The resulting information is used to build the coroutine dispatch switch
// statements.
func trackDispatchSpans(stmt ast.Stmt) map[ast.Stmt]dispatchSpan {
	spans := map[ast.Stmt]dispatchSpan{}
	trackDispatchSpans0(stmt, spans, 1)
	return spans
}

type dispatchSpan struct{ start, end int }

func trackDispatchSpans0(stmt ast.Stmt, dispatchSpans map[ast.Stmt]dispatchSpan, nextID int) int {
	startID := nextID
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		for _, child := range s.List {
			nextID = trackDispatchSpans0(child, dispatchSpans, nextID)
		}
		if len(s.List) == 0 {
			nextID++
		}
	case *ast.IfStmt:
		nextID = trackDispatchSpans0(s.Body, dispatchSpans, nextID)
		if s.Else != nil {
			nextID = trackDispatchSpans0(s.Else, dispatchSpans, nextID)
		}
	case *ast.ForStmt:
		nextID = trackDispatchSpans0(s.Body, dispatchSpans, nextID)
	case *ast.SwitchStmt:
		nextID = trackDispatchSpans0(s.Body, dispatchSpans, nextID)
	case *ast.TypeSwitchStmt:
		nextID = trackDispatchSpans0(s.Body, dispatchSpans, nextID)
	case *ast.CaseClause:
		for _, child := range s.Body {
			nextID = trackDispatchSpans0(child, dispatchSpans, nextID)
		}
	case *ast.SelectStmt:
		nextID = trackDispatchSpans0(s.Body, dispatchSpans, nextID)
	case *ast.CommClause:
		for _, child := range s.Body {
			nextID = trackDispatchSpans0(child, dispatchSpans, nextID)
		}
	case *ast.LabeledStmt:
		nextID = trackDispatchSpans0(s.Stmt, dispatchSpans, nextID)
	default:
		nextID++ // leaf
	}
	dispatchSpans[stmt] = dispatchSpan{startID, nextID}
	return nextID
}

// compileDispatch adds the coroutine's dispatch statements to a tree.
//
// The dispatch mechanism is used when recursively rewinding stacks to
// resume a coroutine. Each function on the stack frame needs to jump
// to the correct location in the code, even when there are arbitrary
// levels of branches and loops. To do this, we generate a switch inside
// each block, using the information from trackDispatchSpans.
func compileDispatch(stmt ast.Stmt, dispatchSpans map[ast.Stmt]dispatchSpan) ast.Stmt {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		switch {
		case len(s.List) == 1:
			child := compileDispatch(s.List[0], dispatchSpans)
			s.List[0] = unnestBlocks(child)
		case len(s.List) > 1:
			stmt = &ast.BlockStmt{List: []ast.Stmt{compileDispatch0(s.List, dispatchSpans)}}
		}
	case *ast.IfStmt:
		s.Body = compileDispatch(s.Body, dispatchSpans).(*ast.BlockStmt)
	case *ast.ForStmt:
		forSpan := dispatchSpans[s]
		s.Body = compileDispatch(s.Body, dispatchSpans).(*ast.BlockStmt)

		// Hijack the loop's post iteration statement to inject an IP reset.
		if s.Post == nil {
			s.Post = &ast.AssignStmt{Lhs: []ast.Expr{}, Tok: token.ASSIGN, Rhs: []ast.Expr{}}
		} else if incDec, ok := s.Post.(*ast.IncDecStmt); ok {
			var op token.Token
			switch incDec.Tok {
			case token.INC:
				op = token.ADD
			case token.DEC:
				op = token.SUB
			}
			s.Post = &ast.AssignStmt{
				Lhs: []ast.Expr{incDec.X},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{&ast.BinaryExpr{X: incDec.X, Op: op, Y: &ast.BasicLit{Kind: token.INT, Value: "1"}}},
			}
		}
		assign, ok := s.Post.(*ast.AssignStmt)
		if !ok {
			panic("not implemented")
		}
		if assign.Tok != token.ASSIGN {
			for i := range assign.Lhs {
				var op token.Token
				switch assign.Tok {
				case token.ADD_ASSIGN:
					op = token.ADD
				case token.SUB_ASSIGN:
					op = token.SUB
				case token.MUL_ASSIGN:
					op = token.MUL
				case token.QUO_ASSIGN:
					op = token.QUO
				case token.REM_ASSIGN:
					op = token.REM
				case token.AND_ASSIGN:
					op = token.AND
				case token.OR_ASSIGN:
					op = token.OR
				case token.XOR_ASSIGN:
					op = token.XOR
				case token.SHL_ASSIGN:
					op = token.SHL
				case token.SHR_ASSIGN:
					op = token.SHR
				case token.AND_NOT_ASSIGN:
					op = token.AND_NOT
				}
				// From the Go language spec:
				// > An assignment operation x op= y where op is a binary arithmetic operator is equivalent to x = x op (y) but evaluates x only once.
				// Thus, this transformation is only valid if the LHS doesn't
				// contain side effects. This is checked elsewhere.
				assign.Rhs[i] = &ast.BinaryExpr{X: assign.Lhs[i], Op: op, Y: assign.Rhs[i]}
			}
			assign.Tok = token.ASSIGN
		}
		assign.Lhs = append(assign.Lhs, &ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")})
		assign.Rhs = append(assign.Rhs, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(forSpan.start)})

	case *ast.SwitchStmt:
		for i, child := range s.Body.List {
			s.Body.List[i] = compileDispatch(child, dispatchSpans)
		}
	case *ast.TypeSwitchStmt:
		for i, child := range s.Body.List {
			s.Body.List[i] = compileDispatch(child, dispatchSpans)
		}
	case *ast.SelectStmt:
		for i, child := range s.Body.List {
			s.Body.List[i] = compileDispatch(child, dispatchSpans)
		}
	case *ast.CaseClause:
		switch {
		case len(s.Body) == 1:
			child := compileDispatch(s.Body[0], dispatchSpans)
			s.Body[0] = unnestBlocks(child)
		case len(s.Body) > 1:
			s.Body = []ast.Stmt{compileDispatch0(s.Body, dispatchSpans)}
		}
	case *ast.CommClause:
		switch {
		case len(s.Body) == 1:
			child := compileDispatch(s.Body[0], dispatchSpans)
			s.Body[0] = unnestBlocks(child)
		case len(s.Body) > 1:
			s.Body = []ast.Stmt{compileDispatch0(s.Body, dispatchSpans)}
		}
	case *ast.LabeledStmt:
		s.Stmt = compileDispatch(s.Stmt, dispatchSpans)
	}
	return stmt
}

func compileDispatch0(stmts []ast.Stmt, dispatchSpans map[ast.Stmt]dispatchSpan) ast.Stmt {
	var cases []ast.Stmt
	for i, child := range stmts {
		childSpan := dispatchSpans[child]
		compiledChild := compileDispatch(child, dispatchSpans)
		compiledChild = unnestBlocks(compiledChild)
		caseBody := []ast.Stmt{compiledChild}
		if i < len(stmts)-1 {
			caseBody = append(caseBody,
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")}},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(childSpan.end)}},
				},
				&ast.BranchStmt{Tok: token.FALLTHROUGH})
		}
		cases = append(cases, &ast.CaseClause{
			List: []ast.Expr{
				&ast.BinaryExpr{
					X:  &ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")},
					Op: token.LSS, /* < */
					Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(childSpan.end)}},
			},
			Body: caseBody,
		})
	}
	return &ast.SwitchStmt{Body: &ast.BlockStmt{List: cases}}
}

func unnestBlocks(stmt ast.Stmt) ast.Stmt {
	for {
		s, ok := stmt.(*ast.BlockStmt)
		if !ok || len(s.List) != 1 {
			return stmt
		}
		stmt = s.List[0]
	}
}
