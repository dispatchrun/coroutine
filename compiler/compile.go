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
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const coroutinePackage = "github.com/stealthrocket/coroutine"

// Compile compiles coroutines in a module.
//
// The path argument can either be a path to a package within
// the module, or a pattern that matches multiple packages in the
// module (for example, /path/to/module/...). In both cases, the
// nearest module is located and compiled as a whole.
//
// The path can be absolute, or relative to the current working directory.
func Compile(path string, options ...Option) error {
	c := &compiler{
		fset: token.NewFileSet(),
	}
	for _, option := range options {
		option(c)
	}
	return c.compile(path)
}

// Option configures the compiler.
type Option func(*compiler)

// WithBuildTags instructs the compiler to attach the specified build
// tags to generated files.
func WithBuildTags(buildTags string) Option {
	return func(c *compiler) { c.buildTags = buildTags }
}

type compiler struct {
	buildTags string

	fset *token.FileSet
}

func (c *compiler) compile(path string) error {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	var dotdotdot bool
	absPath, dotdotdot = strings.CutSuffix(absPath, "...")
	if s, err := os.Stat(absPath); err != nil {
		return err
	} else if !s.IsDir() {
		// Make sure we're loading whole packages.
		absPath = filepath.Dir(absPath)
	}
	var pattern string
	if dotdotdot {
		pattern = "./..."
	} else {
		pattern = "."
	}

	log.Printf("reading, parsing and type-checking")
	conf := &packages.Config{
		Mode: packages.NeedName | packages.NeedModule |
			packages.NeedImports | packages.NeedDeps |
			packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes,
		Fset: c.fset,
		Dir:  absPath,
		Env:  os.Environ(),
	}
	pkgs, err := packages.Load(conf, pattern)
	if err != nil {
		return fmt.Errorf("packages.Load %q: %w", path, err)
	}
	var moduleDir string
	for _, p := range pkgs {
		if p.Module == nil {
			return fmt.Errorf("package %s is not part of a module", p.PkgPath)
		}
		if moduleDir == "" {
			moduleDir = p.Module.Dir
		} else if moduleDir != p.Module.Dir {
			return fmt.Errorf("pattern more than one module (%s + %s)", moduleDir, p.Module.Dir)
		}
	}
	err = nil
	packages.Visit(pkgs, func(p *packages.Package) bool {
		for _, e := range p.Errors {
			err = e
			break
		}
		return err == nil
	}, nil)
	if err != nil {
		return err
	}

	log.Printf("building SSA program")
	prog, _ := ssautil.AllPackages(pkgs, ssa.InstantiateGenerics|ssa.GlobalDebug)
	prog.Build()

	log.Printf("building call graph")
	cg := vta.CallGraph(ssautil.AllFunctions(prog), cha.CallGraph(prog))

	log.Printf("finding generic yield instantiations")
	var coroutinePkg *packages.Package
	packages.Visit(pkgs, func(p *packages.Package) bool {
		if p.PkgPath == coroutinePackage {
			coroutinePkg = p
		}
		return coroutinePkg == nil
	}, nil)
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
	packages.Visit(pkgs, func(p *packages.Package) bool {
		pkgsByTypes[p.Types] = p
		return true
	}, nil)
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

	var needVendoring []*packages.Package
	for p := range colorsByPkg {
		if p.Module == nil || p.Module.Dir != moduleDir {
			needVendoring = append(needVendoring, p)
			break
		}
	}
	if len(needVendoring) > 0 {
		log.Printf("vendoring packages")
		newRoot := filepath.Join(moduleDir, "goroot")
		if err := vendor(newRoot, needVendoring); err != nil {
			return err
		}
	}

	for p, colors := range colorsByPkg {
		if err := c.compilePackage(p, colors); err != nil {
			return err
		}
	}

	log.Printf("done")
	return nil
}

func (c *compiler) writeFile(path string, file *ast.File) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
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
	if _, err := f.WriteString(b.String()); err != nil {
		return err
	}
	// Format/write the remainder of the AST.
	if err := format.Node(f, c.fset, file); err != nil {
		return err
	}
	return f.Close()
}

func (c *compiler) compilePackage(p *packages.Package, colors functionColors) error {
	log.Printf("compiling package %s", p.Name)

	colorsByFunc := map[ast.Node]*types.Signature{}
	for fn, color := range colors {
		decl := fn.Syntax()
		switch decl.(type) {
		case *ast.FuncDecl:
		case *ast.FuncLit:
		default:
			return fmt.Errorf("unsupported yield function %s (Syntax is %T, not *ast.FuncDecl or *ast.FuncLit)", fn, decl)
		}
		colorsByFunc[decl] = color
	}

	for i, f := range p.Syntax {
		// Generate the coroutine AST.
		gen := &ast.File{
			Name: ast.NewIdent(p.Name),
		}

		for _, anydecl := range f.Decls {
			decl, ok := anydecl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			color, ok := colorsByFunc[decl]
			if !ok {
				gen.Decls = append(gen.Decls, decl)
				continue
			}
			// Reject certain language features for now.
			if err := unsupported(decl, p.TypesInfo); err != nil {
				return err
			}

			scope := &scope{colors: colorsByFunc}
			// If the function has a single expression it does not contain a
			// deferred closure; it won't be added to the list of colored
			// functions so generateFunctypes does not mistakenly increment the
			// local symbol counter when generating closure names.
			gen.Decls = append(gen.Decls, scope.compileFuncDecl(p, decl, color))
		}

		if len(gen.Decls) == 0 {
			continue
		}

		generateFunctypes(p, gen, colorsByFunc)

		// Find all the required imports for this file.
		gen = addImports(p, gen)

		outputPath, _ := strings.CutSuffix(p.GoFiles[i], ".go")
		if c.buildTags != "" {
			outputPath += "_" + strings.ReplaceAll(c.buildTags, ",", "_") + ".go"
		} else {
			outputPath += "_generated.go"
		}
		if err := c.writeFile(outputPath, gen); err != nil {
			return err
		}
	}

	return nil
}

func addImports(p *packages.Package, gen *ast.File) *ast.File {
	imports := map[string]string{}

	ast.Inspect(gen, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			ident, ok := x.X.(*ast.Ident)
			if !ok || ident.Name == "" {
				break
			}

			obj := p.TypesInfo.ObjectOf(ident)
			pkgname, ok := obj.(*types.PkgName)
			if !ok {
				break
			}

			pkg := pkgname.Imported().Path()
			if pkg == "" {
				break
			}

			if existing, ok := imports[ident.Name]; ok && existing != pkg {
				fmt.Println("existing:", ident.Name, existing)
				fmt.Println("new:", pkg)
				panic("conflicting imports")
			}
			imports[ident.Name] = pkg
		}
		return true
	})

	if len(imports) == 0 {
		return gen
	}

	importspecs := make([]ast.Spec, 0, len(imports))
	for name, path := range imports {
		importspecs = append(importspecs, &ast.ImportSpec{
			Name: ast.NewIdent(name),
			Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(path)},
		})
	}

	gen.Decls = append([]ast.Decl{&ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: importspecs,
	}}, gen.Decls...)

	return gen
}

type scope struct {
	colors map[ast.Node]*types.Signature
	// Index used to generate unique object identifiers within the scope of a
	// function.
	//
	// The field is reset to zero after compiling function declarations because
	// we don't need globally unique identifiers for local variables.
	//
	// See decls.go for usage.
	objectIndex int
	// Index used to generate unique frame identifiers with the scope of a
	// function.
	//
	// Unique names are necessary to allow closures to reference
	frameIndex int
}

func (scope *scope) compileFuncDecl(p *packages.Package, fn *ast.FuncDecl, color *types.Signature) *ast.FuncDecl {
	log.Printf("compiling function %s %s", p.Name, fn.Name)

	// Generate the coroutine function. At this stage, use the same name
	// as the source function (and require that the caller use build tags
	// to disambiguate function calls).
	gen := &ast.FuncDecl{
		Doc:  &ast.CommentGroup{},
		Name: fn.Name,
		Type: funcTypeWithNamedResults(fn.Type),
		Body: scope.compileFuncBody(p, fn.Type, fn.Body, color),
	}

	// If the function declaration contains function literals, we have to
	// add the //go:noinline copmiler directive to prevent inlining or the
	// resulting symbol name generated by the linker wouldn't match the
	// predictions made in generateFunctypes.
	//
	// When functions are inlined, the linker creates a unique name
	// combining the symbol name of the calling function and the symbol name
	// of the closure. Knowing which functions will be inlined is difficult
	// considering the score-base mechansim that Go uses and alterations
	// like PGO, therefore we take the simple approach of disabling inlining
	// instead.
	//
	// Note that we only need to do this for single-expression functions as
	// otherwise the presence of a defer statement to unwind the coroutine
	// already prevents inlining, however, it's simpler to always add the
	// compiler directive.
	gen.Doc.List = appendCommentGroup(gen.Doc.List, fn.Doc)
	gen.Doc.List = appendComment(gen.Doc.List, "//go:noinline\n")

	if !isExpr(gen.Body) {
		scope.colors[gen] = color
	}
	return gen
}

func (scope *scope) compileFuncLit(p *packages.Package, fn *ast.FuncLit, color *types.Signature) *ast.FuncLit {
	log.Printf("compiling function literal %s", p.Name)

	gen := &ast.FuncLit{
		Type: funcTypeWithNamedResults(fn.Type),
		Body: scope.compileFuncBody(p, fn.Type, fn.Body, color),
	}

	if !isExpr(gen.Body) {
		scope.colors[gen] = color
	}
	return gen
}

func (scope *scope) compileFuncBody(p *packages.Package, typ *ast.FuncType, body *ast.BlockStmt, color *types.Signature) *ast.BlockStmt {
	var defers *ast.Ident

	mayYield := findCalls(body, p.TypesInfo)
	markBranchStmt(body, mayYield)

	body = desugar(p, body, mayYield).(*ast.BlockStmt)
	body = astutil.Apply(body,
		func(cursor *astutil.Cursor) bool {
			switch n := cursor.Node().(type) {
			case *ast.FuncLit:
				color, ok := scope.colors[n]
				if ok {
					cursor.Replace(scope.compileFuncLit(p, n, color))
				}
				return false
			case *ast.DeferStmt:
				if defers == nil {
					// This identifier is created to represent the local
					// variable collecting defers but it gets rewritten to
					// use a field on the stack frame so the list of defers
					// can be captured by the coroutine.
					defers = ast.NewIdent("_defers")
					p.TypesInfo.Defs[defers] = types.NewVar(0, p.Types, defers.Name,
						types.NewSlice(types.NewSignatureType(nil, nil, nil, nil, nil, false)),
					)
				}
				cursor.Replace(&ast.AssignStmt{
					Lhs: []ast.Expr{defers},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun:  ast.NewIdent("append"),
							Args: []ast.Expr{defers, n.Call.Fun},
						},
					},
				})
			}
			return true
		},
		nil,
	).(*ast.BlockStmt)

	if isExpr(body) {
		return body
	}

	gen := new(ast.BlockStmt)
	ctx := ast.NewIdent("_c")
	frame := ast.NewIdent("_f")
	fp := ast.NewIdent("_fp")

	yieldTypeExpr := make([]ast.Expr, 2)
	yieldTypeExpr[0] = typeExpr(p, color.Params().At(0).Type())
	yieldTypeExpr[1] = typeExpr(p, color.Results().At(0).Type())

	// _c := coroutine.LoadContext[R, S]()
	gen.List = append(gen.List, &ast.AssignStmt{
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

	// _f, _fp := _c.Push()
	gen.List = append(gen.List, &ast.AssignStmt{
		Lhs: []ast.Expr{frame, fp},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Push")},
			},
		},
	})

	frameName := ast.NewIdent(fmt.Sprintf("_f%d", scope.frameIndex))
	scope.frameIndex++

	// Handle declarations.
	//
	// Types, constants and variables can be defined within any scope in the
	// function, and can shadow previous declarations. The coroutine dispatch
	// mechanism introduces new scopes, which may prevent the declarations from
	// being visible to other statements, or may cause some statements to
	// unexpectedly observe an unshadowed type or value.
	//
	// To handle shadowing, we assign each type, constant and variable a unique
	// name within the function body. To handle scoping issues, we hoist
	// declarations to the function prologue. We downgrade inline var decls and
	// assignments that use := to assignments that use =. Constant decls are
	// hoisted and also have their value assigned in the function prologue.
	decls, frameType, frameInit := extractDecls(p, typ, body, defers, p.TypesInfo)
	renameObjects(body, p.TypesInfo, decls, frameName, frameType, frameInit, scope)

	for _, decl := range decls {
		gen.List = append(gen.List, &ast.DeclStmt{Decl: decl})
	}

	gen.List = append(gen.List,
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{frameName},
						Type:  &ast.StarExpr{X: frameType},
					},
				},
			},
		},
	)

	gen.List = append(gen.List, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")},
			Op: token.EQL, /* == */
			Y:  &ast.BasicLit{Kind: token.INT, Value: "0"}},
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{frameName},
			Rhs: []ast.Expr{&ast.UnaryExpr{Op: token.AND, X: frameInit}},
		}}},
		Else: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{frameName},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.TypeAssertExpr{
				X: &ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: frame, Sel: ast.NewIdent("Get")},
					Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
				Type: &ast.StarExpr{X: frameType},
			}},
		}}},
	})

	popFrame := []ast.Stmt{
		&ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Pop")}}},
	}

	if defers != nil {
		popFrame = append(popFrame, &ast.RangeStmt{
			Key:   ast.NewIdent("_"),
			Value: ast.NewIdent("f"),
			Tok:   token.DEFINE,
			X: &ast.SelectorExpr{
				X:   frameName,
				Sel: frameType.Fields.List[len(frameType.Fields.List)-1].Names[0],
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.DeferStmt{Call: &ast.CallExpr{Fun: ast.NewIdent("f")}},
			}},
		})
	}

	gen.List = append(gen.List, &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.FuncLit{
				Type: &ast.FuncType{Params: new(ast.FieldList)},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.IfStmt{
							Cond: &ast.CallExpr{
								Fun: &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Unwinding")},
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.ExprStmt{X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{X: frame, Sel: ast.NewIdent("Set")},
										Args: []ast.Expr{
											&ast.BasicLit{Kind: token.INT, Value: "0"},
											frameName,
										},
									}},
									&ast.ExprStmt{X: &ast.CallExpr{
										Fun:  &ast.SelectorExpr{X: ctx, Sel: ast.NewIdent("Store")},
										Args: []ast.Expr{fp, frame},
									}},
								},
							},
							Else: &ast.BlockStmt{List: popFrame},
						},
					},
				},
			},
		},
	})

	spans := trackDispatchSpans(body)
	mayYield = findCalls(body, p.TypesInfo)
	compiledBody := compileDispatch(body, spans, mayYield).(*ast.BlockStmt)
	gen.List = append(gen.List, compiledBody.List...)

	// If the function returns one or more values, it must end with a return statement;
	// we inject it if the function body does not already has one.
	if typ.Results != nil && len(typ.Results.List) > 0 {
		needsReturn := len(gen.List) == 0
		if !needsReturn {
			_, endsWithReturn := gen.List[len(gen.List)-1].(*ast.ReturnStmt)
			needsReturn = !endsWithReturn
		}
		if needsReturn {
			gen.List = append(gen.List, &ast.ReturnStmt{})
		}
	}

	return gen
}

// This function returns true if a function body is composed of at most one
// expression.
func isExpr(body *ast.BlockStmt) bool {
	if len(body.List) == 0 {
		return true
	}
	if len(body.List) == 1 {
		if _, isExpr := body.List[0].(*ast.ExprStmt); isExpr {
			return true
		}
	}
	return false
}

func funcTypeWithNamedResults(t *ast.FuncType) *ast.FuncType {
	if t.Results == nil {
		return t
	}
	underscore := ast.NewIdent("_")
	funcType := *t
	funcType.Results = &ast.FieldList{
		List: slices.Clone(t.Results.List),
	}
	for i, f := range t.Results.List {
		if len(f.Names) == 0 {
			field := *f
			field.Names = []*ast.Ident{underscore}
			funcType.Results.List[i] = &field
		}
	}
	return &funcType
}
