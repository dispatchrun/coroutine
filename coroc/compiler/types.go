package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
)

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
