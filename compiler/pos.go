package compiler

import (
	"go/ast"
	"go/token"
)

// clearPos resets the token.Pos field(s) in each type of ast.Node.
// When AST nodes are generated alongside nodes that have position
// information, it can cause formatting to produce invalid results.
// This function clears the information from all nodes so that the
// formatter produces correct results.
func clearPos(tree ast.Node) {
	ast.Inspect(tree, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ArrayType:
			n.Lbrack = token.NoPos
		case *ast.AssignStmt:
			n.TokPos = token.NoPos
		case *ast.BasicLit:
			n.ValuePos = token.NoPos
		case *ast.BinaryExpr:
			n.OpPos = token.NoPos
		case *ast.BlockStmt:
			n.Rbrace = token.NoPos
			n.Lbrace = token.NoPos
		case *ast.BranchStmt:
			n.TokPos = token.NoPos
		case *ast.CallExpr:
			n.Ellipsis = token.NoPos
			n.Lparen = token.NoPos
			n.Rparen = token.NoPos
		case *ast.CaseClause:
			n.Colon = token.NoPos
			n.Case = token.NoPos
		case *ast.ChanType:
			n.Arrow = token.NoPos
			n.Begin = token.NoPos
		case *ast.CommClause:
			n.Case = token.NoPos
			n.Colon = token.NoPos
		case *ast.Comment:
			n.Slash = token.NoPos
		case *ast.CommentGroup:
		case *ast.CompositeLit:
			n.Lbrace = token.NoPos
			n.Rbrace = token.NoPos
		case *ast.DeclStmt:
		case *ast.DeferStmt:
			n.Defer = token.NoPos
		case *ast.Ellipsis:
			n.Ellipsis = token.NoPos
		case *ast.EmptyStmt:
			n.Semicolon = token.NoPos
		case *ast.ExprStmt:
		case *ast.Field:
		case *ast.FieldList:
			n.Closing = token.NoPos
			n.Opening = token.NoPos
		case *ast.File:
			n.Package = token.NoPos
			n.FileStart = token.NoPos
			n.FileEnd = token.NoPos
		case *ast.ForStmt:
			n.For = token.NoPos
		case *ast.FuncDecl:
		case *ast.FuncLit:
		case *ast.FuncType:
			n.Func = token.NoPos
		case *ast.GenDecl:
			n.Lparen = token.NoPos
			n.Rparen = token.NoPos
			n.TokPos = token.NoPos
		case *ast.GoStmt:
			n.Go = token.NoPos
		case *ast.Ident:
			n.NamePos = token.NoPos
		case *ast.IfStmt:
			n.If = token.NoPos
		case *ast.ImportSpec:
			n.EndPos = token.NoPos
		case *ast.IncDecStmt:
			n.TokPos = token.NoPos
		case *ast.IndexExpr:
			n.Lbrack = token.NoPos
			n.Rbrack = token.NoPos
		case *ast.IndexListExpr:
			n.Lbrack = token.NoPos
			n.Rbrack = token.NoPos
		case *ast.InterfaceType:
			n.Interface = token.NoPos
		case *ast.KeyValueExpr:
			n.Colon = token.NoPos
		case *ast.LabeledStmt:
			n.Colon = token.NoPos
		case *ast.MapType:
			n.Map = token.NoPos
		case *ast.Package:
		case *ast.ParenExpr:
			n.Lparen = token.NoPos
			n.Rparen = token.NoPos
		case *ast.RangeStmt:
			n.TokPos = token.NoPos
			n.For = token.NoPos
			n.Range = token.NoPos
		case *ast.ReturnStmt:
			n.Return = token.NoPos
		case *ast.SelectStmt:
			n.Select = token.NoPos
		case *ast.SelectorExpr:
		case *ast.SendStmt:
			n.Arrow = token.NoPos
		case *ast.SliceExpr:
			n.Lbrack = token.NoPos
			n.Rbrack = token.NoPos
		case *ast.StarExpr:
			n.Star = token.NoPos
		case *ast.StructType:
			n.Struct = token.NoPos
		case *ast.SwitchStmt:
			n.Switch = token.NoPos
		case *ast.TypeAssertExpr:
			n.Lparen = token.NoPos
			n.Rparen = token.NoPos
		case *ast.TypeSpec:
			n.Assign = token.NoPos
		case *ast.TypeSwitchStmt:
			n.Switch = token.NoPos
		case *ast.UnaryExpr:
			n.OpPos = token.NoPos
		case *ast.ValueSpec:
		}
		return true
	})
}
