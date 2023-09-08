package compiler

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

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

	// Load, parse and type-check packages and their dependencies.
	conf := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedImports | packages.NeedDeps | packages.NeedTypesInfo,
		Fset: c.fset,
	}
	pkgs, err := packages.Load(conf, path)
	if err != nil {
		return fmt.Errorf("packages.Load %q: %w", path, err)
	}
	for _, p := range pkgs {
		// Keep it simple and only return the first error (the goal is to
		// compile coroutines, not worry about error reporting UX).
		for _, err := range p.Errors {
			return err
		}
	}

	// At this stage, candidate packages are those that import the
	// coroutine package and explicitly yield. This could be relaxed
	// in future so that candidate packages are those that *may*
	// contain a yield point.
	for _, p := range pkgs {
		if _, ok := p.Imports[coroutinePackage]; !ok {
			continue
		}
		if err := c.compilePackage(p); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) compilePackage(p *packages.Package) error {
	// At this stage, candidate functions are those that explicitly
	// yield. This could be relaxed in future so that candidate packages
	// are those that *may* yield, or those that are explicitly
	// whitelisted by the user (either via command-line opt-in, or by
	// annotating the function with some comment directive).
	type coroutineCandidate struct {
		FuncDecl  *ast.FuncDecl
		YieldType []ast.Expr
	}
	var candidates []coroutineCandidate
	for _, f := range p.Syntax {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			// Skip function declarations without a body.
			if fn.Body == nil {
				continue
			}
			// Skip methods (functions with receivers) for now.
			if fn.Recv != nil {
				continue
			}
			var yieldType []ast.Expr
			scanYields(p, fn.Body, func(t []ast.Expr) bool {
				if yieldType == nil {
					yieldType = t
				} else {
					// TODO: fail if t isn't the same as yieldType (i.e. more than one type of yield here)
				}
				return true
			})
			if yieldType == nil {
				continue
			}
			candidates = append(candidates, coroutineCandidate{
				FuncDecl:  fn,
				YieldType: yieldType,
			})
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	// Reject certain language features for now.
	for _, candidate := range candidates {
		fn := candidate.FuncDecl

		var err error
		ast.Inspect(fn, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.DeferStmt:
				err = fmt.Errorf("not implemented: defer")
			case *ast.GoStmt:
				err = fmt.Errorf("not implemented: go")
			case *ast.LabeledStmt:
				err = fmt.Errorf("not implemented: labels")
			case *ast.TypeSwitchStmt:
				err = fmt.Errorf("not implemented: type switch")
			case *ast.SelectStmt:
				err = fmt.Errorf("not implemented: select")
			case *ast.RangeStmt:
				err = fmt.Errorf("not implemented: for range")
			case *ast.DeclStmt:
				err = fmt.Errorf("not implemented: inline decls")
			case *ast.AssignStmt:
				if len(n.Lhs) != 1 || len(n.Lhs) != len(n.Rhs) {
					err = fmt.Errorf("not implemented: multiple assign")
				}
				if _, ok := n.Lhs[0].(*ast.Ident); !ok {
					err = fmt.Errorf("not implemented: assign to non-ident")
				}
			case *ast.BranchStmt:
				if n.Tok == token.GOTO {
					err = fmt.Errorf("not implemented: goto")
				} else if n.Tok == token.FALLTHROUGH {
					err = fmt.Errorf("not implemented: fallthrough")
				} else if n.Tok == token.BREAK {
					err = fmt.Errorf("not implemented: break")
				} else if n.Tok == token.CONTINUE {
					err = fmt.Errorf("not implemented: continue")
				} else if n.Label != nil {
					err = fmt.Errorf("not implemented: labeled branch")
				}
			case *ast.ForStmt:
				// Since we aren't desugaring for loop post iteration
				// statements yet, check that it's a simple increment
				// or decrement.
				switch p := n.Post.(type) {
				case nil:
				case *ast.IncDecStmt:
					if _, ok := p.X.(*ast.Ident); !ok {
						err = fmt.Errorf("not implemented: for post inc/dec %T", p.X)
					}
				default:
					err = fmt.Errorf("not implemented: for post %T", p)
				}
			}
			return err == nil
		})
		if err != nil {
			return err
		}

		// Require int params/return values for now.
		if fn.Type.Params != nil {
			for _, fn := range fn.Type.Params.List {
				if ident, ok := fn.Type.(*ast.Ident); !ok || ident.Name != "int" {
					return fmt.Errorf("not implemented: non-int params")
				}
			}
		}
		if fn.Type.Results != nil {
			for _, fn := range fn.Type.Results.List {
				if ident, ok := fn.Type.(*ast.Ident); !ok || ident.Name != "int" {
					return fmt.Errorf("not implemented: non-int results")
				}
			}
		}
	}

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
	for _, candidate := range candidates {
		fn, yieldType := candidate.FuncDecl, candidate.YieldType
		gen.Decls = append(gen.Decls, c.compileFunction(p, fn, yieldType))
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

func (c *compiler) compileFunction(p *packages.Package, fn *ast.FuncDecl, yieldType []ast.Expr) *ast.FuncDecl {
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
					Indices: yieldType,
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

	// Scan/replace variables defined in the function.
	objectVars := map[*ast.Object]*ast.Ident{}
	var varNames []*ast.Ident
	var varTypes []types.Type
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			name := n.Lhs[0].(*ast.Ident)
			if n.Tok == token.DEFINE {
				n.Tok = token.ASSIGN
			}
			if name.Obj == nil {
				return true
			}
			if _, ok := objectVars[name.Obj]; ok {
				return true
			}
			varName := ast.NewIdent("_v" + strconv.Itoa(len(varNames)))
			varTypes = append(varTypes, p.TypesInfo.TypeOf(name))
			varNames = append(varNames, varName)
			objectVars[name.Obj] = varName
		}
		return true
	})
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			if replacement, ok := objectVars[ident.Obj]; ok {
				ident.Name = replacement.Name
			}
		}
		return true
	})

	// Declare variables upfront.
	if len(varTypes) > 0 {
		varDecl := &ast.GenDecl{Tok: token.VAR}
		for i, t := range varTypes {
			var typeExpr ast.Expr
			if t.String() == "int" {
				typeExpr = ast.NewIdent("int")
			} else {
				// TODO: types.Type => ast.Expr
				panic("not implemented")
			}
			varDecl.Specs = append(varDecl.Specs, &ast.ValueSpec{
				Names: []*ast.Ident{varNames[i]},
				Type:  typeExpr,
			})
		}
		gen.Body.List = append(gen.Body.List, &ast.DeclStmt{Decl: varDecl})
	}

	// Collect params/results/variables that need to be saved/restored.
	var saveAndRestore []*ast.Ident
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				saveAndRestore = append(saveAndRestore, name)
			}
		}
	}
	if fn.Type.Results != nil {
		// Named results could be used as scratch space at any point
		// during execution, so they need to be saved/restored.
		for _, result := range fn.Type.Results.List {
			for _, name := range result.Names {
				saveAndRestore = append(saveAndRestore, name)
			}
		}
	}
	saveAndRestore = append(saveAndRestore, varNames...)

	// Restore state when rewinding the stack.
	//
	// As an optimization, only those variables still in scope for a
	// particular f.IP need to be restored.
	var restoreStmts []ast.Stmt
	for i, name := range saveAndRestore {
		restoreStmts = append(restoreStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{name},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: ast.NewIdent("int"),
					Args: []ast.Expr{
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
							Type: &ast.SelectorExpr{X: ast.NewIdent("coroutine"), Sel: ast.NewIdent("Int")},
						},
					},
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
	for i, name := range saveAndRestore {
		saveStmts = append(saveStmts, &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: frame, Sel: ast.NewIdent("Set")},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(i)},
					&ast.CallExpr{
						Fun:  &ast.SelectorExpr{X: ast.NewIdent("coroutine"), Sel: ast.NewIdent("Int")},
						Args: []ast.Expr{name},
					},
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
						},
					},
				},
			},
		},
	})

	desugar(fn.Body)

	spans := trackSpans(fn.Body)

	gen.Body.List = append(gen.Body.List, c.compileStatement(fn.Body, spans).(*ast.BlockStmt).List...)

	return gen
}

func (c *compiler) compileStatement(stmt ast.Stmt, spans map[ast.Stmt]span) ast.Stmt {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		switch {
		case len(s.List) == 1:
			s.List[0] = c.compileStatement(s.List[0], spans)
		case len(s.List) > 1:
			s = &ast.BlockStmt{List: []ast.Stmt{c.compileDispatch(s.List, spans)}}
		}
		return s
	case *ast.IfStmt:
		s.Body = c.compileStatement(s.Body, spans).(*ast.BlockStmt)
		return s
	case *ast.ForStmt:
		forSpan := spans[s]
		s.Body = c.compileStatement(s.Body, spans).(*ast.BlockStmt)
		s.Body.List = append(s.Body.List, &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")}},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(forSpan.start)}},
		})
		return s
	case *ast.SwitchStmt:
		for i, child := range s.Body.List {
			s.Body.List[i] = c.compileStatement(child, spans)
		}
		return s
	case *ast.CaseClause:
		switch {
		case len(s.Body) == 1:
			s.Body[0] = c.compileStatement(s.Body[0], spans)
		case len(s.Body) > 1:
			s.Body = []ast.Stmt{c.compileDispatch(s.Body, spans)}
		}
		return s
	default:
		return s
	}
	return stmt
}

func (c *compiler) compileDispatch(stmts []ast.Stmt, spans map[ast.Stmt]span) ast.Stmt {
	var cases []ast.Stmt
	for i, child := range stmts {
		childSpan := spans[child]
		caseBody := []ast.Stmt{c.compileStatement(child, spans)}
		if i < len(stmts)-1 {
			caseBody = append(caseBody,
				&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")}},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(childSpan.end)}},
				},
				&ast.BranchStmt{Tok: token.FALLTHROUGH})
		}
		cases = append(cases, &ast.CaseClause{
			List: []ast.Expr{
				&ast.BinaryExpr{
					X:  &ast.SelectorExpr{X: ast.NewIdent("_f"), Sel: ast.NewIdent("IP")},
					Op: token.LSS, /* < */
					Y:  &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(childSpan.end)}},
			},
			Body: caseBody,
		})
	}
	return &ast.SwitchStmt{Body: &ast.BlockStmt{List: cases}}
}
