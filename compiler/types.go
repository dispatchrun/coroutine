package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/packages"
)

// typeExpr converts a types.Type to an ast.Expr.
//
// If typeArg is provided, it's used to resolve type parameters.
func typeExpr(p *packages.Package, typ types.Type, typeArg func(*types.TypeParam) types.Type) ast.Expr {
	switch t := typ.(type) {
	case *types.Basic:
		switch t {
		case types.Typ[types.UntypedBool]:
			t = types.Typ[types.Bool]
		}
		return ast.NewIdent(t.String())
	case *types.Slice:
		return &ast.ArrayType{Elt: typeExpr(p, t.Elem(), typeArg)}
	case *types.Array:
		return &ast.ArrayType{
			Len: &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(t.Len(), 10)},
			Elt: typeExpr(p, t.Elem(), typeArg),
		}
	case *types.Map:
		return &ast.MapType{
			Key:   typeExpr(p, t.Key(), typeArg),
			Value: typeExpr(p, t.Elem(), typeArg),
		}
	case *types.Struct:
		fields := make([]*ast.Field, t.NumFields())
		for i := range fields {
			f := t.Field(i)
			fields[i] = &ast.Field{Type: typeExpr(p, f.Type(), typeArg)}
			if !f.Anonymous() {
				fields[i].Names = []*ast.Ident{ast.NewIdent(f.Name())}
			}
			if tag := t.Tag(i); tag != "" {
				panic("not implemented: struct tags")
			}
		}
		return &ast.StructType{Fields: &ast.FieldList{List: fields}}
	case *types.Pointer:
		return &ast.StarExpr{X: typeExpr(p, t.Elem(), typeArg)}
	case *types.Interface:
		if t.Empty() {
			return ast.NewIdent("any")
		}
	case *types.Signature:
		return newFuncType(p, t)
	case *types.Named:
		obj := t.Obj()
		name := ast.NewIdent(obj.Name())
		pkg := obj.Pkg()

		var namedExpr ast.Expr
		if pkg == nil || p.Types == pkg {
			namedExpr = name
		} else {
			// Update the package's type map to track that this package is
			// imported with this identifier. We do not attempt to reuse
			// identifiers at the moment.
			pkgident := ast.NewIdent(pkg.Name())
			p.TypesInfo.Uses[pkgident] = types.NewPkgName(pkgident.NamePos, p.Types, pkgident.Name, pkg)
			namedExpr = &ast.SelectorExpr{X: pkgident, Sel: name}
		}
		if typeArgs := t.TypeArgs(); typeArgs != nil {
			indices := make([]ast.Expr, typeArgs.Len())
			for i := range indices {
				indices[i] = typeExpr(p, typeArgs.At(i), typeArg)
			}
			namedExpr = &ast.IndexListExpr{
				X:       namedExpr,
				Indices: indices,
			}
		}
		return namedExpr

	case *types.Chan:
		c := &ast.ChanType{
			Value: typeExpr(p, t.Elem(), typeArg),
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

	case *types.TypeParam:
		if typeArg != nil {
			return typeExpr(p, typeArg(t), typeArg)
		}
		obj := t.Obj()
		ident := ast.NewIdent(obj.Name())
		p.TypesInfo.Defs[ident] = obj
		return ident
	}
	panic(fmt.Sprintf("not implemented: %T", typ))
}

func newFuncType(p *packages.Package, signature *types.Signature) *ast.FuncType {
	return &ast.FuncType{
		Params:  newFieldList(p, signature.Params()),
		Results: newFieldList(p, signature.Results()),
	}
}

func newFieldList(p *packages.Package, tuple *types.Tuple) *ast.FieldList {
	return &ast.FieldList{
		List: newFields(p, tuple),
	}
}

func newFields(p *packages.Package, tuple *types.Tuple) []*ast.Field {
	fields := make([]*ast.Field, tuple.Len())
	for i := range fields {
		fields[i] = &ast.Field{
			Type: typeExpr(p, tuple.At(i).Type(), nil),
		}
	}
	return fields
}
