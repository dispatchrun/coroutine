package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// unsupported checks a tree for unsupported language features.
func unsupported(tree ast.Node, info *types.Info) (err error) {
	ast.Inspect(tree, func(node ast.Node) bool {
		stmt, ok := node.(ast.Stmt)
		if !ok {
			return true
		}
		switch n := stmt.(type) {
		// Not yet supported:
		case *ast.DeferStmt:
			err = fmt.Errorf("not implemented: defer")
		case *ast.GoStmt:
			err = fmt.Errorf("not implemented: go")
		case *ast.LabeledStmt:
			err = fmt.Errorf("not implemented: labels")
		case *ast.TypeSwitchStmt:
			err = fmt.Errorf("not implemented: type switch")
		case *ast.SelectStmt:
			err = fmt.Errorf("not implemented: select")
		case *ast.CommClause:
			err = fmt.Errorf("not implemented: select case")
		case *ast.DeclStmt:
			err = fmt.Errorf("not implemented: inline decls")

		// Partially supported:
		case *ast.RangeStmt:
			switch t := info.TypeOf(n.X).(type) {
			case *types.Array, *types.Slice:
			default:
				err = fmt.Errorf("not implemented: for range for %T", t)
			}
		case *ast.AssignStmt:
			if len(n.Lhs) != 1 || len(n.Lhs) != len(n.Rhs) {
				err = fmt.Errorf("not implemented: multiple assign")
			}
			if _, ok := n.Lhs[0].(*ast.Ident); !ok {
				err = fmt.Errorf("not implemented: assign to non-ident")
			}
		case *ast.BranchStmt:
			if n.Tok == token.GOTO {
				err = fmt.Errorf("not implemented: goto")
			} else if n.Tok == token.FALLTHROUGH {
				err = fmt.Errorf("not implemented: fallthrough")
			} else if n.Tok == token.BREAK {
				err = fmt.Errorf("not implemented: break")
			} else if n.Tok == token.CONTINUE {
				err = fmt.Errorf("not implemented: continue")
			} else if n.Label != nil {
				err = fmt.Errorf("not implemented: labeled branch")
			}
		case *ast.ForStmt:
			// Since we aren't desugaring for loop post iteration
			// statements yet, check that it's a simple increment
			// or decrement.
			switch p := n.Post.(type) {
			case nil:
			case *ast.IncDecStmt:
				if _, ok := p.X.(*ast.Ident); !ok {
					err = fmt.Errorf("not implemented: for post inc/dec %T", p.X)
				}
			default:
				err = fmt.Errorf("not implemented: for post %T", p)
			}

		// Fully supported:
		case *ast.BlockStmt:
		case *ast.CaseClause:
		case *ast.EmptyStmt:
		case *ast.ExprStmt:
		case *ast.IfStmt:
		case *ast.IncDecStmt:
		case *ast.ReturnStmt:
		case *ast.SendStmt:
		case *ast.SwitchStmt:

		// Catch all in case new statements are added:
		default:
			err = fmt.Errorf("not implmemented: ast.Stmt(%T)", n)
		}
		return err == nil
	})
	return
}
