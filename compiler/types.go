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
		return newFuncType(p, t, typeArg)
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

func newFuncType(p *packages.Package, signature *types.Signature, typeArg func(*types.TypeParam) types.Type) *ast.FuncType {
	return &ast.FuncType{
		Params:  newFieldList(p, signature.Params(), typeArg),
		Results: newFieldList(p, signature.Results(), typeArg),
	}
}

func newFieldList(p *packages.Package, tuple *types.Tuple, typeArg func(*types.TypeParam) types.Type) *ast.FieldList {
	return &ast.FieldList{
		List: newFields(p, tuple, typeArg),
	}
}

func newFields(p *packages.Package, tuple *types.Tuple, typeArg func(*types.TypeParam) types.Type) []*ast.Field {
	fields := make([]*ast.Field, tuple.Len())
	for i := range fields {
		fields[i] = &ast.Field{
			Type: typeExpr(p, tuple.At(i).Type(), typeArg),
		}
	}
	return fields
}

// substituteTypeArgs replaces all type parameter placeholders
// with type args.
//
// It returns a deep copy of the input expr.
func substituteTypeArgs(p *packages.Package, expr ast.Expr, typeArg func(*types.TypeParam) types.Type) ast.Expr {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.ArrayType:
		return &ast.ArrayType{
			Elt: substituteTypeArgs(p, e.Elt, typeArg),
			Len: substituteTypeArgs(p, e.Len, typeArg),
		}
	case *ast.MapType:
		return &ast.MapType{
			Key:   substituteTypeArgs(p, e.Key, typeArg),
			Value: substituteTypeArgs(p, e.Value, typeArg),
		}
	case *ast.FuncType:
		return &ast.FuncType{
			TypeParams: substituteFieldList(p, e.TypeParams, typeArg),
			Params:     substituteFieldList(p, e.Params, typeArg),
			Results:    substituteFieldList(p, e.Results, typeArg),
		}
	case *ast.ChanType:
		return &ast.ChanType{
			Dir:   e.Dir,
			Value: substituteTypeArgs(p, e.Value, typeArg),
		}
	case *ast.StructType:
		return &ast.StructType{
			Fields: substituteFieldList(p, e.Fields, typeArg),
		}
	case *ast.StarExpr:
		return &ast.StarExpr{
			X: substituteTypeArgs(p, e.X, typeArg),
		}
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{
			X:   substituteTypeArgs(p, e.X, typeArg),
			Sel: e.Sel,
		}
	case *ast.IndexExpr:
		return &ast.IndexExpr{
			X:     substituteTypeArgs(p, e.X, typeArg),
			Index: substituteTypeArgs(p, e.Index, typeArg),
		}
	case *ast.Ident:
		t := p.TypesInfo.TypeOf(e)
		tp, ok := t.(*types.TypeParam)
		if !ok {
			return e
		}
		return typeExpr(p, typeArg(tp), typeArg)
	case *ast.BasicLit:
		return e
	default:
		panic(fmt.Sprintf("not implemented: %T", e))
	}
}

func substituteFieldList(p *packages.Package, f *ast.FieldList, typeArg func(*types.TypeParam) types.Type) *ast.FieldList {
	if f == nil || f.List == nil {
		return f
	}
	fields := make([]*ast.Field, len(f.List))
	for i, field := range f.List {
		fields[i] = &ast.Field{
			Names: field.Names,
			Type:  substituteTypeArgs(p, field.Type, typeArg),
			Tag:   field.Tag,
		}
	}
	return &ast.FieldList{List: fields}
}
