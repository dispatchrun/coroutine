package compiler

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const coroutinePackage = "github.com/stealthrocket/coroutine"

// Compile compiles coroutines in one or more packages.
//
// The path argument can either be a path to a package, a
// path to a file within a package, or a pattern that matches
// multiple packages (for example, /path/to/package/...).
// The path can be absolute or relative (to the current working
// directory).
func Compile(path string, options ...CompileOption) error {
	c := &compiler{
		outputFilename: "coroc_generated.go",
		fset:           token.NewFileSet(),
	}
	for _, option := range options {
		option(c)
	}
	return c.compile(path)
}

// CompileOption configures the compiler.
type CompileOption func(*compiler)

// WithOutputFilename instructs the compiler to write generated code
// to a file with the specified name within each package that contains
// coroutines.
func WithOutputFilename(outputFilename string) CompileOption {
	return func(c *compiler) { c.outputFilename = outputFilename }
}

// WithBuildTags instructs the compiler to attach the specified build
// tags to generated files.
func WithBuildTags(buildTags string) CompileOption {
	return func(c *compiler) { c.buildTags = buildTags }
}

type compiler struct {
	outputFilename string
	buildTags      string

	fset *token.FileSet
}

func (c *compiler) compile(path string) error {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if path != "" && !strings.HasSuffix(path, "...") {
		s, err := os.Stat(path)
		if err != nil {
			return err
		} else if !s.IsDir() {
			// Make sure we're loading whole packages.
			path = filepath.Dir(path)
		}
	}
	path = filepath.Clean(path)
	if len(path) > 0 && path[0] != filepath.Separator && path[0] != '.' {
		// Go interprets patterns without a leading dot as part of the
		// stdlib (i.e. part of $GOROOT/src) rather than relative to
		// the working dir. Note that filepath.Join(".", path) does not
		// give the desired result here, hence the manual concat.
		path = "." + string(filepath.Separator) + path
	}

	log.Printf("reading, parsing and type-checking")
	conf := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedTypesInfo,
		Fset: c.fset,
	}

	pkgs, err := packages.Load(conf, path)
	if err != nil {
		return fmt.Errorf("packages.Load %q: %w", path, err)
	}
	flatpkgs := flattenPackages(pkgs)
	for _, p := range flatpkgs {
		for _, err := range p.Errors {
			return err
		}
	}

	log.Printf("building SSA program")
	prog, _ := ssautil.Packages(pkgs, ssa.InstantiateGenerics|ssa.GlobalDebug)
	prog.Build()

	log.Printf("building call graph")
	cg := cha.CallGraph(prog)

	log.Printf("finding generic yield instantiations")
	var coroutinePkg *packages.Package
	for _, p := range flatpkgs {
		if p.PkgPath == coroutinePackage {
			coroutinePkg = p
			break
		}
	}
	if coroutinePkg == nil {
		log.Printf("%s not imported by the module. Nothing to do", coroutinePackage)
		return nil
	}
	yieldFunc := prog.FuncValue(coroutinePkg.Types.Scope().Lookup("Yield").(*types.Func))
	yieldInstances := functionColors{}
	for fn := range ssautil.AllFunctions(prog) {
		if fn.Origin() == yieldFunc {
			yieldInstances[fn] = fn.Signature
		}
	}

	log.Printf("coloring functions")
	colors, err := colorFunctions(cg, yieldInstances)
	if err != nil {
		return err
	}
	pkgsByTypes := map[*types.Package]*packages.Package{}
	for _, p := range flatpkgs {
		pkgsByTypes[p.Types] = p
	}
	colorsByPkg := map[*packages.Package]functionColors{}
	for fn, color := range colors {
		if fn.Pkg == nil {
			return fmt.Errorf("unsupported yield function %s (Pkg is nil)", fn)
		}

		p := pkgsByTypes[fn.Pkg.Pkg]
		pkgColors := colorsByPkg[p]
		if pkgColors == nil {
			pkgColors = functionColors{}
			colorsByPkg[p] = pkgColors
		}
		pkgColors[fn] = color
	}

	for p, colors := range colorsByPkg {
		if err := c.compilePackage(p, colors); err != nil {
			return err
		}
	}

	log.Printf("done")

	return nil
}

func (c *compiler) compilePackage(p *packages.Package, colors functionColors) error {
	log.Printf("compiling package %s", p.Name)

	// Generate the coroutine AST.
	gen := &ast.File{
		Name: ast.NewIdent(p.Name),
	}
	gen.Decls = append(gen.Decls, &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(coroutinePackage)},
			},
		},
	})

	colorsByDecl := map[*ast.FuncDecl]*types.Signature{}
	for fn, color := range colors {
		decl, ok := fn.Syntax().(*ast.FuncDecl)
		if !ok {
			return fmt.Errorf("unsupported yield function %s (Syntax is %T, not *ast.FuncDecl)", fn, fn.Syntax())
		}
		colorsByDecl[decl] = color
	}
	for _, f := range p.Syntax {
		for _, anydecl := range f.Decls {
			decl, ok := anydecl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			color, ok := colorsByDecl[decl]
			if !ok {
				continue
			}

			// Reject certain language features for now.
			if err := unsupported(decl, p.TypesInfo); err != nil {
				return err
			}

			gen.Decls = append(gen.Decls, c.compileFunction(p, decl, color))
		}
	}

	log.Print("building type register init function")
	if err := generateTypesInit(c.fset, gen, p); err != nil {
		return err
	}

	// Get ready to write.
	packageDir := filepath.Dir(p.GoFiles[0])
	outputPath := filepath.Join(packageDir, c.outputFilename)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("os.Create %q: %w", outputPath, err)
	}
	defer outputFile.Close()

	// Comments are awkward to attach to the tree (they rely on token.Pos, which
	// is coupled to a token.FileSet). Instead, just write out the raw strings.
	var b strings.Builder
	b.WriteString(`// Code generated by coroc. DO NOT EDIT`)
	b.WriteString("\n\n")
	if c.buildTags != "" {
		b.WriteString(`//go:build `)
		b.WriteString(c.buildTags)
		b.WriteString("\n\n")
	}
	if _, err := outputFile.WriteString(b.String()); err != nil {
		return err
	}

	// Format/write the remainder of the AST.
	if err := format.Node(outputFile, c.fset, gen); err != nil {
		return err
	}

	return outputFile.Close()
}

func (c *compiler) compileFunction(p *packages.Package, fn *ast.FuncDecl, color *types.Signature) *ast.FuncDecl {
	log.Printf("compiling function %s %s", p.Name, fn.Name)

	// Generate the coroutine function. At this stage, use the same name
	// as the source function (and require that the caller use build tags
	// to disambiguate function calls).
	gen := &ast.FuncDecl{
		Name: fn.Name,
		Type: fn.Type,
		Body: &ast.BlockStmt{},
	}

	ctx := ast.NewIdent("_c")
	frame := ast.NewIdent("_f")

	yieldTypeExpr := make([]ast.Expr, 2)
	yieldTypeExpr[0] = typeExpr(color.Params().At(0).Type())
	yieldTypeExpr[1] = typeExpr(color.Results().At(0).Type())

	// _c := coroutine.LoadContext[R, S]()
	gen.Body.List = append(gen.Body.List, &ast.AssignStmt{
		Lhs: []ast.Expr{ctx},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.IndexListExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent("coroutine"),
						Sel: ast.NewIdent("LoadContext"),
					},
					Indices: yieldTypeExpr,
				},
			},
		},
	})

	// _f := _c.Push()
	gen.Body.List = append(gen.Body.List, &ast.AssignStmt{
		Lhs: []ast.Expr{frame},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Push")},
			},
		},
	})

	// Desugar statements in the tree.
	desugar(fn.Body, p.TypesInfo)

	// Scan/replace variables defined in the function.
	//
	// Variable declarations are moved to the function prologue so that
	// variables can be saved and restored. To handle cases of shadowing,
	// all variables are given new unique names of the form _v[0-9]+.
	// Inline declarations (via var or :=) are downgraded to an assignment
	// using =.
	objectVars := map[types.Object]*ast.Ident{}
	var varNames []*ast.Ident
	var varTypes []types.Type
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				// Rewrite := to = here, since it doesn't require an AST
				// node replacement.
				n.Tok = token.ASSIGN
			}
			for _, lhs := range n.Lhs {
				name := lhs.(*ast.Ident)
				obj := p.TypesInfo.ObjectOf(name)
				if obj == nil {
					return true
				}
				if _, ok := objectVars[obj]; ok {
					return true
				}
				varName := ast.NewIdent("_v" + strconv.Itoa(len(varNames)))
				varTypes = append(varTypes, p.TypesInfo.TypeOf(name))
				varNames = append(varNames, varName)
				objectVars[obj] = varName
			}
		case *ast.ValueSpec:
			// Rewrite var decls in a pass below, since it does require an AST
			// node replacement.
			for _, name := range n.Names {
				obj := p.TypesInfo.ObjectOf(name)
				if obj == nil {
					return true
				}
				if _, ok := objectVars[obj]; ok {
					return true
				}
				varName := ast.NewIdent("_v" + strconv.Itoa(len(varNames)))
				varTypes = append(varTypes, p.TypesInfo.TypeOf(name))
				varNames = append(varNames, varName)
				objectVars[obj] = varName
			}
		}
		return true
	})
	astutil.Apply(fn.Body, func(cursor *astutil.Cursor) bool {
		declStmt, ok := cursor.Node().(*ast.DeclStmt)
		if !ok {
			return true
		}
		g, ok := declStmt.Decl.(*ast.GenDecl)
		if !ok {
			return true
		}
		if !ok || g.Tok != token.VAR {
			return true
		}
		var assigns []ast.Stmt
		// The var decl could have one spec, e.g. var foo=0, or multiple
		// specs, e.g. var ( foo=0; bar=1; baz=2 ). Replace them with a
		// block that has one or more assignments. Pure decls can be omitted.
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
		cursor.Replace(&ast.BlockStmt{List: assigns})
		return true
	}, nil)
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			obj := p.TypesInfo.ObjectOf(ident)
			if replacement, ok := objectVars[obj]; ok {
				ident.Name = replacement.Name
			}
		}
		return true
	})

	// Declare variables upfront.
	if len(varTypes) > 0 {
		varDecl := &ast.GenDecl{Tok: token.VAR}
		for i, t := range varTypes {
			varDecl.Specs = append(varDecl.Specs, &ast.ValueSpec{
				Names: []*ast.Ident{varNames[i]},
				Type:  typeExpr(t),
			})
		}
		gen.Body.List = append(gen.Body.List, &ast.DeclStmt{Decl: varDecl})
	}

	// Collect params/results/variables that need to be saved/restored.
	var saveAndRestoreNames []*ast.Ident
	var saveAndRestoreTypes []types.Type
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				if name.Name != "_" {
					saveAndRestoreNames = append(saveAndRestoreNames, name)
					saveAndRestoreTypes = append(saveAndRestoreTypes, p.TypesInfo.TypeOf(name))
				}
			}
		}
	}
	if fn.Type.Results != nil {
		// Named results could be used as scratch space at any point
		// during execution, so they need to be saved/restored.
		for _, result := range fn.Type.Results.List {
			for _, name := range result.Names {
				if name.Name != "_" {
					saveAndRestoreNames = append(saveAndRestoreNames, name)
					saveAndRestoreTypes = append(saveAndRestoreTypes, p.TypesInfo.TypeOf(name))
				}
			}
		}
	}
	saveAndRestoreNames = append(saveAndRestoreNames, varNames...)
	saveAndRestoreTypes = append(saveAndRestoreTypes, varTypes...)

	// Restore state when rewinding the stack.
	//
	// As an optimization, only those variables still in scope for a
	// particular f.IP need to be restored.
	var restoreStmts []ast.Stmt
	for i, name := range saveAndRestoreNames {
		restoreStmts = append(restoreStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{name},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.TypeAssertExpr{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   frame,
							Sel: ast.NewIdent("Get"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i)},
						},
					},
					Type: typeExpr(saveAndRestoreTypes[i]),
				},
			},
		})
	}
	gen.Body.List = append(gen.Body.List, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")},
			Op: token.GTR, /* > */
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"}},
		Body: &ast.BlockStmt{List: restoreStmts},
	})

	// Save state when unwinding the stack.
	var saveStmts []ast.Stmt
	for i, name := range saveAndRestoreNames {
		saveStmts = append(saveStmts, &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: frame, Sel: ast.NewIdent("Set")},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i)},
					name,
				},
			},
		})
	}
	gen.Body.List = append(gen.Body.List, &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Type: &ast.FuncType{},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.IfStmt{
							Cond: &ast.CallExpr{
								Fun: &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Unwinding")},
							},
							Body: &ast.BlockStmt{List: saveStmts},
							Else: &ast.BlockStmt{List: []ast.Stmt{
								&ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Pop")}}}},
							},
						},
					},
				},
			},
		},
	})

	spans := trackDispatchSpans(fn.Body)

	compiledBody := compileDispatch(fn.Body, spans).(*ast.BlockStmt)

	gen.Body.List = append(gen.Body.List, compiledBody.List...)

	return gen
}
