package compiler

import (
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
func extractDecls(tree ast.Node, info *types.Info) (decls []*ast.GenDecl) {
	ast.Inspect(tree, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			// Stop when we encounter a function listeral so we don't hoist its
			// local variables into the scope of its parent function.
			return false
		case *ast.GenDecl: // const, var, type
			if n.Tok == token.TYPE || n.Tok == token.CONST {
				decls = append(decls, n)
			} else {
				// Var specs can have multiple values on the left and/or right
				// hand side, and the types may vary (e.g. var wd, err = os.Getwd()).
				// Since types may vary in the same spec, just generate a new decl
				// for each declared name.
				for _, spec := range n.Specs {
					s := spec.(*ast.ValueSpec)
					for _, name := range s.Names {
						if name.Name == "_" {
							continue
						}
						decls = append(decls, &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{name},
									Type:  typeExpr(info.TypeOf(name)),
								},
							},
						})
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
				decls = append(decls, &ast.GenDecl{
					Tok: token.VAR,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{name},
							Type:  typeExpr(t),
						},
					},
				})
			}
		}
		return true
	})
	return
}

// renameObjects renames types, constants and variables declared within
// a function. Each is given a unique name, so that declarations are safe
// to hoist into the function prologue.
func renameObjects(tree ast.Node, info *types.Info, decls []*ast.GenDecl, scope *scope) {
	// Scan decls to find objects, giving each new object a unique name.
	newNames := map[types.Object]*ast.Ident{}
	for _, decl := range decls {
		for _, spec := range decl.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec: // type
				if s.Name.Name == "_" {
					continue
				}
				newNames[info.ObjectOf(s.Name)] = scope.newObjectIdent()
			case *ast.ValueSpec: // const/var
				for _, name := range s.Names {
					if name.Name == "_" {
						continue
					}
					newNames[info.ObjectOf(name)] = scope.newObjectIdent()
				}
			}
		}
	}
	// Add type info for the new identifiers.
	for obj, name := range newNames {
		info.Defs[name] = types.NewVar(0, nil, name.Name, obj.Type())
	}
	// Rename identifiers in the tree.
	ast.Inspect(tree, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			if replacement, ok := newNames[info.ObjectOf(ident)]; ok {
				ident.Name = replacement.Name
			}
		}
		return true
	})
}

// removeDecls removes type, constant and variable declarations from a tree.
// Variable declarations via assignment (:=) are instead downgraded to =.
func removeDecls(tree ast.Node) {
	astutil.Apply(tree, func(cursor *astutil.Cursor) bool {
		switch n := cursor.Node().(type) {
		case *ast.FuncLit:
			return false
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				if _, ok := cursor.Parent().(*ast.TypeSwitchStmt); ok {
					return true // preserve type switch decls.
				}
				n.Tok = token.ASSIGN // otherwise, convert := to =
			}
		case *ast.DeclStmt:
			g, ok := n.Decl.(*ast.GenDecl)
			if !ok {
				return true
			}
			switch g.Tok {
			case token.TYPE, token.CONST:
				// Delete type and const decls, since they'll be hoisted to the
				// function prologue.
				cursor.Delete()
			case token.VAR:
				// The var decl could have one spec, e.g. var foo=0, or
				// multiple specs, e.g. var ( foo=0; bar=1; baz=2 ). Some
				// specs may have values and type and some might not, e.g.
				// var (foo int; bar = 1; baz int = 2). Remove the pure
				// decls and the decl assignments with types, leaving only
				// assignments, e.g. { bar = 1; baz = 2 }
				var assigns []ast.Stmt
				for _, spec := range g.Specs {
					s, ok := spec.(*ast.ValueSpec)
					if !ok || len(s.Values) == 0 {
						continue
					}
					lhs := make([]ast.Expr, len(s.Names))
					for i, name := range s.Names {
						lhs[i] = name
					}
					assigns = append(assigns, &ast.AssignStmt{Lhs: lhs, Tok: token.ASSIGN, Rhs: s.Values})
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
		}
		return true
	}, nil)
}

func scanDeclVarIdentifiers(decls []*ast.GenDecl, fn func(*ast.Ident)) {
	for _, decl := range decls {
		if decl.Tok != token.VAR {
			continue
		}
		for _, spec := range decl.Specs {
			v, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range v.Names {
				if name.Name != "_" {
					fn(name)
				}
			}
		}
	}
}

func scanFuncTypeIdentifiers(t *ast.FuncType, fn func(*ast.Ident)) {
	if t.Params != nil {
		for _, param := range t.Params.List {
			for _, name := range param.Names {
				if name.Name != "_" {
					fn(name)
				}
			}
		}
	}
	if t.Results != nil {
		for _, result := range t.Results.List {
			for _, name := range result.Names {
				if name.Name != "_" {
					fn(name)
				}
			}
		}
	}
}
