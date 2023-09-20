package compiler

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestDesugar(t *testing.T) {
	for _, test := range []struct {
		name   string
		body   string
		expect string
		uses   map[string]types.Object
		defs   map[string]types.Object
		types  map[string]types.TypeAndValue
	}{
		{
			name: "if cond",
			body: "if bar == 1 {}",
			expect: `
{
	_v0 := bar == 1
	if _v0 {
	}
}`,
		},
		{
			name: "if init + cond",
			body: "if foo := bar; bar == 1 {}",
			expect: `
{
	foo := bar
	_v0 := bar == 1
	if _v0 {
	}
}`,
		},
		{
			name: "for init + cond + post",
			body: "for i := 0; i < 10; i++ { result += i }",
			expect: `
{
	i := 0
_l0:
	for ; ; i++ {
		{
			_v0 := !(i < 10)
			if _v0 {
				break _l0
			}
		}
		result += i
	}
}`,
		},
		{
			name: "labeled for",
			body: "outer: for i := 0; i < 10; i++ { for j := 0; j < 10; j++ { break outer } }",
			expect: `
{
	i := 0
_l0:
	for ; ; i++ {
		{
			_v0 := !(i < 10)
			if _v0 {
				break _l0
			}
		}
		{
			j := 0
		_l1:
			for ; ; j++ {
				{
					_v1 := !(j < 10)
					if _v1 {
						break _l1
					}
				}
				break _l0
			}
		}
	}
}`,
		},
		{
			name: "labeled for break and continue handling",
			body: `
outer:
	for {
		switch {
		case true:
			break
		case false:
			continue
		default:
			break outer
		}
	}`,
			expect: `
_l0:
	for {
	_l1:
		switch {
		case true:
			break _l1
		case false:
			continue _l0
		default:
			break _l0
		}
	}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			expr, err := parser.ParseExpr("func() {\n" + test.body + "\n}()")
			if err != nil {
				t.Fatal(err)
			}
			body := expr.(*ast.CallExpr).Fun.(*ast.FuncLit).Body

			info := &types.Info{
				Defs:  map[*ast.Ident]types.Object{},
				Uses:  map[*ast.Ident]types.Object{},
				Types: map[ast.Expr]types.TypeAndValue{},
			}
			ast.Inspect(body, func(node ast.Node) bool {
				if ident, ok := node.(*ast.Ident); ok {
					if obj, ok := test.defs[ident.Name]; ok {
						info.Defs[ident] = obj
					}
					if obj, ok := test.uses[ident.Name]; ok {
						info.Uses[ident] = obj
					}
					if t, ok := test.types[ident.Name]; ok {
						info.Types[ident] = t
					}
				}
				return true
			})
			desugared := desugar(body, info)
			desugared = unnestBlocks(desugared)

			expect := strings.TrimSpace(test.expect)
			actual := formatNode(desugared)
			if actual != expect {
				t.Errorf("unexpected desugared result")
				t.Logf("expect:\n%s", test.expect)
				t.Logf("actual:\n%s", actual)
			}
		})
	}
}

func formatNode(node ast.Node) string {
	fset := token.NewFileSet()
	//ast.Print(fset, node)
	var b bytes.Buffer
	if err := format.Node(&b, fset, node); err != nil {
		panic(err)
	}
	return b.String()
}
