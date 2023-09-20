package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
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
	case *types.Map:
		return &ast.MapType{
			Key:   typeExpr(t.Key()),
			Value: typeExpr(t.Elem()),
		}
	case *types.Struct:
		fields := make([]*ast.Field, t.NumFields())
		for i := range fields {
			f := t.Field(i)
			fields[i] = &ast.Field{Type: typeExpr(f.Type())}
			if !f.Anonymous() {
				fields[i].Names = []*ast.Ident{ast.NewIdent(f.Name())}
			}
			if tag := t.Tag(i); tag != "" {
				panic("not implemented: struct tags")
			}
		}
		return &ast.StructType{Fields: &ast.FieldList{List: fields}}
	case *types.Pointer:
		return &ast.StarExpr{X: typeExpr(t.Elem())}
	case *types.Interface:
		if t.Empty() {
			return ast.NewIdent("any")
		}
	case *types.Signature:
		return newFuncType(t)
	case *types.Named:
		if t.TypeParams() != nil || t.TypeArgs() != nil {
			panic("not implemented: generic types")
		}
		obj := t.Obj()
		name := ast.NewIdent(obj.Name())
		pkg := obj.Pkg()
		if pkg == nil {
			return name
		}
		// TODO: this needs to be incorporated in the pass to find imports
		return &ast.SelectorExpr{X: ast.NewIdent(pkg.Name()), Sel: name}
	case *types.Chan:
		t.Dir()
		c := &ast.ChanType{
			Value: typeExpr(t.Elem()),
		}
		switch t.Dir() {
		case types.SendRecv:
			c.Dir = ast.SEND | ast.RECV
		case types.SendOnly:
			c.Dir = ast.SEND
		case types.RecvOnly:
			c.Dir = ast.RECV
		}
		return c
	}
	panic(fmt.Sprintf("not implemented: %T", typ))
}

func newFuncType(signature *types.Signature) *ast.FuncType {
	return &ast.FuncType{
		Params:  newFieldList(signature.Params()),
		Results: newFieldList(signature.Results()),
	}
}

func newFieldList(tuple *types.Tuple) *ast.FieldList {
	return &ast.FieldList{
		List: newFields(tuple),
	}
}

func newFields(tuple *types.Tuple) []*ast.Field {
	fields := make([]*ast.Field, tuple.Len())
	for i := range fields {
		fields[i] = &ast.Field{
			Type: typeExpr(tuple.At(i).Type()),
		}
	}
	return fields
}

func funcTypeWithNamedResults(t *ast.FuncType) *ast.FuncType {
	if t.Results == nil {
		return t
	}
	funcType := *t
	funcType.Results = &ast.FieldList{
		List: slices.Clone(t.Results.List),
	}
	for i, f := range t.Results.List {
		if len(f.Names) == 0 {
			field := *f
			field.Names = []*ast.Ident{
				ast.NewIdent("_"),
			}
			funcType.Results.List[i] = &field
		}
	}
	return &funcType
}
