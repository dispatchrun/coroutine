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
				fields[i].Tag = &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(tag)}
			}
		}
		return &ast.StructType{Fields: &ast.FieldList{List: fields}}
	case *types.Pointer:
		return &ast.StarExpr{X: typeExpr(p, t.Elem(), typeArg)}
	case *types.Interface:
		if t.Empty() {
			return ast.NewIdent("any")
		}
		methods := make([]*ast.Field, t.NumExplicitMethods())
		for i := range methods {
			m := t.ExplicitMethod(i)
			methods[i] = &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(m.Name())},
				Type:  typeExpr(p, m.Type(), typeArg),
			}
		}
		embeddeds := make([]*ast.Field, t.NumEmbeddeds())
		for i := range embeddeds {
			embeddeds[i] = &ast.Field{
				Type: typeExpr(p, t.EmbeddedType(i), typeArg),
			}
		}
		return &ast.InterfaceType{
			Methods: &ast.FieldList{
				List: append(methods, embeddeds...),
			},
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
			if known := typeArg(t); known != nil {
				return typeExpr(p, known, typeArg)
			}
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
	case *ast.InterfaceType:
		return &ast.InterfaceType{
			Methods: substituteFieldList(p, e.Methods, typeArg),
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
	case *ast.IndexListExpr:
		indices := make([]ast.Expr, len(e.Indices))
		for i, index := range e.Indices {
			indices[i] = substituteTypeArgs(p, index, typeArg)
		}
		return &ast.IndexListExpr{
			X:       substituteTypeArgs(p, e.X, typeArg),
			Indices: indices,
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

func containsTypeParam(typ types.Type) bool {
	if typ == nil {
		return false
	}
	switch t := typ.(type) {
	case *types.Basic:
	case *types.Slice:
		return containsTypeParam(t.Elem())
	case *types.Array:
		return containsTypeParam(t.Elem())
	case *types.Pointer:
		return containsTypeParam(t.Elem())
	case *types.Map:
		return containsTypeParam(t.Elem()) || containsTypeParam(t.Key())
	case *types.Chan:
		return containsTypeParam(t.Elem())
	case *types.Named:
		if args := t.TypeArgs(); args != nil {
			for i := 0; i < args.Len(); i++ {
				if containsTypeParam(args.At(i)) {
					return true
				}
			}
		}
	case *types.Tuple:
		for i := 0; i < t.Len(); i++ {
			if containsTypeParam(t.At(i).Type()) {
				return true
			}
		}
	case *types.Signature:
		if recv := t.Recv(); recv != nil {
			if containsTypeParam(recv.Type()) {
				return true
			}
		}
		if containsTypeParam(t.Params()) {
			return true
		}
		if containsTypeParam(t.Results()) {
			return true
		}
	case *types.TypeParam:
		return true
	case *types.Interface:
	case *types.Struct:
	default:
		panic(fmt.Sprintf("not implemented: %T", typ))
	}
	return false
}
