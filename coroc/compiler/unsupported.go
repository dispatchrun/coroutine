package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// unsupported checks a function for unsupported language features.
func unsupported(decl *ast.FuncDecl, info *types.Info) (err error) {
	ast.Inspect(decl, func(node ast.Node) bool {
		switch nn := node.(type) {
		case ast.Expr:
			switch nn.(type) {
			case *ast.FuncLit:
				err = fmt.Errorf("not implemented: func literals")
			}
		case ast.Stmt:
			switch n := nn.(type) {
			// Not yet supported:
			case *ast.DeferStmt:
				err = fmt.Errorf("not implemented: defer")
			case *ast.GoStmt:
				err = fmt.Errorf("not implemented: go")
			case *ast.SelectStmt:
				err = fmt.Errorf("not implemented: select")
			case *ast.CommClause:
				err = fmt.Errorf("not implemented: select case")

			// Partially supported:
			case *ast.RangeStmt:
				switch t := info.TypeOf(n.X).(type) {
				case *types.Array, *types.Slice:
				default:
					err = fmt.Errorf("not implemented: for range for %T", t) // e.g. *types.Map
				}
			case *ast.BranchStmt:
				if n.Tok == token.GOTO {
					err = fmt.Errorf("not implemented: goto")
				} else if n.Tok == token.FALLTHROUGH {
					err = fmt.Errorf("not implemented: fallthrough")
				}
			case *ast.LabeledStmt:
				switch n.Stmt.(type) {
				case *ast.ForStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
				default:
					err = fmt.Errorf("not implemented: labels not attached to for/switch/select")
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
			case *ast.AssignStmt:
			case *ast.BlockStmt:
			case *ast.CaseClause:
			case *ast.DeclStmt:
			case *ast.EmptyStmt:
			case *ast.ExprStmt:
			case *ast.IfStmt:
			case *ast.IncDecStmt:
			case *ast.ReturnStmt:
			case *ast.SendStmt:
			case *ast.SwitchStmt:
			case *ast.TypeSwitchStmt:

			// Catch all in case new statements are added:
			default:
				err = fmt.Errorf("not implmemented: ast.Stmt(%T)", n)
			}
		}
		return err == nil
	})
	return
}
