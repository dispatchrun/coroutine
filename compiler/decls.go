package compiler

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
)

// extractDecls extracts type, constant and variable declarations
// from a function body.
//
// Variable declarations via var and via := assignments are included, but
// only the name and type (not the value).
//
// The declaration order is preserved in case types refer to constants and vice
// versa.
//
// Note that declarations are extracted from all nested scopes within the
// function body, so there may be duplicate identifiers. Identifiers can be
// disambiguated using (*types.Info).ObjectOf(ident).
func extractDecls(p *types.Package, typ *ast.FuncType, body *ast.BlockStmt, info *types.Info) (decls []*ast.GenDecl, frameType *ast.StructType, frameInit *ast.CompositeLit) {
	frameType = &ast.StructType{Fields: &ast.FieldList{}}
	frameInit = &ast.CompositeLit{Type: frameType}

	if typ.Params != nil {
		for _, field := range typ.Params.List {
			for _, ident := range field.Names {
				if ident.Name != "_" {
					frameType.Fields.List = append(frameType.Fields.List, &ast.Field{
						Names: []*ast.Ident{ident},
						Type:  field.Type,
					})
					frameInit.Elts = append(frameInit.Elts, &ast.KeyValueExpr{
						Key:   ident,
						Value: ident,
					})
				}
			}
		}
	}

	if typ.Results != nil {
		for _, field := range typ.Results.List {
			for _, ident := range field.Names {
				if ident.Name != "_" {
					frameType.Fields.List = append(frameType.Fields.List, &ast.Field{
						Names: []*ast.Ident{ident},
						Type:  field.Type,
					})
				}
			}
		}
	}

	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			// Stop when we encounter a function listeral so we don't hoist its
			// local variables into the scope of its parent function.
			return false

		case *ast.GenDecl: // const, var, type
			if n.Tok == token.TYPE || n.Tok == token.CONST {
				decls = append(decls, n)
			} else {
				for _, spec := range n.Specs {
					valueSpec := spec.(*ast.ValueSpec)
					valueType := typeExpr(p, info.TypeOf(valueSpec.Names[0]))
					for _, ident := range valueSpec.Names {
						if ident.Name != "_" {
							frameType.Fields.List = append(frameType.Fields.List, &ast.Field{
								Names: []*ast.Ident{ident},
								Type:  valueType,
							})
						}
					}
				}
			}

		case *ast.AssignStmt:
			if n.Tok != token.DEFINE { // := only (not =)
				return true
			}
			for _, lhs := range n.Lhs {
				name := lhs.(*ast.Ident)
				if name.Name == "_" {
					continue
				}
				t := info.TypeOf(lhs)
				if t == nil {
					// Do not hoist the decl in this case. This happens when
					// type switching, e.g.
					//
					//          v-----------v
					//   switch x := y.(type) { ... }
					//
					// The type of x varies depending on the switch case, and
					// has a nil (undefined) type when inspecting the
					// AssignStmt that declares it.
					continue
				}
				frameType.Fields.List = append(frameType.Fields.List, &ast.Field{
					Names: []*ast.Ident{name},
					Type:  typeExpr(p, t),
				})
			}
		}
		return true
	})

	return decls, frameType, frameInit
}

// renameObjects renames types, constants and variables declared within
// a function. Each is given a unique name, so that declarations are safe
// to hoist into the function prologue.
func renameObjects(tree ast.Node, info *types.Info, decls []*ast.GenDecl, frameName *ast.Ident, frameType *ast.StructType, frameInit *ast.CompositeLit, scope *scope) {
	// Scan decls to find objects, giving each new object a unique name.
	names := make(map[types.Object]*ast.Ident, len(decls))
	selectors := make(map[types.Object]*ast.SelectorExpr, len(frameType.Fields.List))

	generateUniqueIdent := func() *ast.Ident {
		ident := scope.objectIndex
		scope.objectIndex++
		return ast.NewIdent(fmt.Sprintf("_o%d", ident))
	}

	addName := func(ident *ast.Ident) {
		if ident.Name != "_" {
			obj := info.ObjectOf(ident)
			newIdent := generateUniqueIdent()
			names[obj] = newIdent
			// Add type info for the new identifiers.
			info.Defs[newIdent] = types.NewVar(0, nil, ident.Name, obj.Type())
		}
	}

	for _, decl := range decls {
		for _, spec := range decl.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec: // type
				addName(s.Name)
			case *ast.ValueSpec: // const/var
				for _, name := range s.Names {
					addName(name)
				}
			}
		}
	}

	frameInitKeyValueExprs := make(map[*ast.Ident]*ast.KeyValueExpr, len(frameInit.Elts))
	for _, elt := range frameInit.Elts {
		expr := elt.(*ast.KeyValueExpr)
		frameInitKeyValueExprs[expr.Key.(*ast.Ident)] = expr
	}

	index := 0
	for i, field := range frameType.Fields.List {
		fieldNames := make([]*ast.Ident, len(field.Names))

		for j, ident := range field.Names {
			if ident.Name != "_" {
				obj := info.ObjectOf(ident)

				newIdent := ast.NewIdent(fmt.Sprintf("X%d", index))
				fieldNames[j] = newIdent
				selectors[obj] = &ast.SelectorExpr{
					X:   frameName,
					Sel: newIdent,
				}

				if expr, ok := frameInitKeyValueExprs[ident]; ok {
					expr.Key = newIdent
				}

				index++
			}
		}

		frameType.Fields.List[i] = &ast.Field{
			Names: fieldNames,
			Type:  field.Type,
		}
	}

	// Once we generated all the new identifers and selectors, we do three
	// passes over the tree:
	//
	// 1. Convert all variable declarations into regular assignments
	// 2. Replace all instances of previous identifiers
	// 3. Remove old const and type declarations
	//
	// The sequence of those operations is important here; we must start with
	// converting the declarations because otherwise we cannot replace nodes of
	// concrete *ast.Ident types with *ast.SelectorExpr, we must turn all
	// declarations into expressions.
	//
	// The removal of constants and types must be done after replacing
	// identifiers, otherwise we might miss some of the identifiers that need
	// replacing if they are removed from the tree too early.
	//
	// Note that replacing identifiers is a recursive operation which traverses
	// function literls.

	astutil.Apply(tree,
		func(cursor *astutil.Cursor) bool {
			switch n := cursor.Node().(type) {
			case *ast.FuncLit:
				return false
			case *ast.DeclStmt:
				switch decl := n.Decl.(*ast.GenDecl); decl.Tok {
				case token.VAR:
					// The var decl could have one spec, e.g. var foo=0, or
					// multiple specs, e.g. var ( foo=0; bar=1; baz=2 ). Some
					// specs may have values and type and some might not, e.g.
					// var (foo int; bar = 1; baz int = 2). Remove the pure
					// decls and the decl assignments with types, leaving only
					// assignments, e.g. { bar = 1; baz = 2 }
					var assigns []ast.Stmt
					for _, spec := range decl.Specs {
						s, ok := spec.(*ast.ValueSpec)
						if !ok || len(s.Values) == 0 {
							continue
						}
						lhs := make([]ast.Expr, len(s.Names))
						for i, name := range s.Names {
							lhs[i] = name
						}
						assigns = append(assigns, &ast.AssignStmt{
							Tok: token.ASSIGN,
							Lhs: lhs,
							Rhs: s.Values,
						})
					}
					switch len(assigns) {
					case 0:
						cursor.Replace(&ast.EmptyStmt{})
					case 1:
						cursor.Replace(assigns[0])
					default:
						cursor.Replace(&ast.BlockStmt{List: assigns})
					}
				}
			case *ast.AssignStmt:
				if n.Tok == token.DEFINE {
					if _, ok := cursor.Parent().(*ast.TypeSwitchStmt); ok {
						return true // preserve type switch decls.
					}
					n.Tok = token.ASSIGN // otherwise, convert := to =
				}
			}
			return true
		},
		nil,
	)

	astutil.Apply(tree,
		func(cursor *astutil.Cursor) bool {
			switch n := cursor.Node().(type) {
			case *ast.Ident:
				obj := info.ObjectOf(n)

				if selector, ok := selectors[obj]; ok {
					cursor.Replace(selector)
				} else if ident, ok := names[obj]; ok {
					cursor.Replace(ident)
				}
			}
			return true
		},
		nil,
	)

	astutil.Apply(tree,
		func(cursor *astutil.Cursor) bool {
			switch n := cursor.Node().(type) {
			case *ast.FuncLit:
				return false
			case *ast.DeclStmt:
				switch decl := n.Decl.(*ast.GenDecl); decl.Tok {
				case token.TYPE, token.CONST:
					// Delete type and const decls, since they'll be hoisted to the
					// function prologue.
					cursor.Delete()
				}
			}
			return true
		},
		nil,
	)
}
