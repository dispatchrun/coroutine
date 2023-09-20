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

	"golang.org/x/tools/go/ast/astutil"
)

func TestDesugar(t *testing.T) {
	intType := types.Typ[types.Int]

	for _, test := range []struct {
		name   string
		body   string
		expect string

		// For simple statements, provide a way to specify type info
		// based on identifier names.
		uses  map[string]types.Object
		defs  map[string]types.Object
		types map[string]types.TypeAndValue
		// For more complex trees, provide a way to access generated
		// *ast.Ident's and modify the type info.
		info func([]ast.Stmt, *types.Info)
	}{
		{
			name: "if cond",
			body: "if bar == 1 { foo }",
			expect: `
{
	_v0 := bar == 1
	if _v0 {
		foo
	}
}`,
		},
		{
			name: "if init + cond",
			body: "if foo := bar; bar == 1 { foo }",
			expect: `
{
	foo := bar
	_v0 := bar == 1
	if _v0 {
		foo
	}
}`,
		},
		{
			name: "if else with init chain",
			body: "if a := 1; a == 1 { foo } else if b := 2; b == 2 { bar } else if c := 3; c == 3 { baz } else { qux }",
			expect: `
{
	a := 1
	_v0 := a == 1
	if _v0 {
		foo
	} else {
		b := 2
		_v1 := b == 2
		if _v1 {
			bar
		} else {
			c := 3
			_v2 := c == 3
			if _v2 {
				baz
			} else {
				qux
			}
		}
	}
}
`,
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
		{
			name: "for range over slice (no index/value)",
			body: "for range []int{0, 1, 2} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		_v1 := 0
	_l0:
		for ; ; _v1++ {
			{
				_v2 := !(_v1 < len(_v0))
				if _v2 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over slice (underscore index)",
			body: "for _ := range []int{0, 1, 2} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		_v1 := 0
	_l0:
		for ; ; _v1++ {
			{
				_v2 := !(_v1 < len(_v0))
				if _v2 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over slice (underscore index and underscore value)",
			body: "for _, _ := range []int{0, 1, 2} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		_v1 := 0
	_l0:
		for ; ; _v1++ {
			{
				_v2 := !(_v1 < len(_v0))
				if _v2 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over slice (index only)",
			body: "for i := range []int{0, 1, 2} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		i := 0
	_l0:
		for ; ; i++ {
			{
				_v1 := !(i < len(_v0))
				if _v1 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over slice (index and underscore value)",
			body: "for i, _ := range []int{0, 1, 2} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		i := 0
	_l0:
		for ; ; i++ {
			{
				_v1 := !(i < len(_v0))
				if _v1 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over slice (index and value)",
			body: "for i, v := range []int{0, 1, 2} {}",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		i := 0
	_l0:
		for ; ; i++ {
			{
				_v1 := !(i < len(_v0))
				if _v1 {
					break _l0
				}
			}
			v := _v0[i]
		}
	}
}
`,
		},
		{
			name: "for range over slice (underscore index and value)",
			body: "for _, v := range []int{0, 1, 2} {}",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewSlice(intType)}
			},
			expect: `
{
	_v0 := []int{0, 1, 2}
	{
		_v1 := 0
	_l0:
		for ; ; _v1++ {
			{
				_v2 := !(_v1 < len(_v0))
				if _v2 {
					break _l0
				}
			}
			v := _v0[_v1]
		}
	}
}
`,
		},
		{
			name: "for range over array (index and value)",
			body: "for i, v := range [3]int{0, 1, 2} {}",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewArray(intType, 3)}
			},
			expect: `
{
	_v0 := [3]int{0, 1, 2}
	{
		i := 0
	_l0:
		for ; ; i++ {
			{
				_v1 := !(i < len(_v0))
				if _v1 {
					break _l0
				}
			}
			v := _v0[i]
		}
	}
}
`,
		},
		{
			name: "for range over map (no index/value)",
			body: "for range map[int]int{} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewMap(intType, intType)}
			},
			expect: `
{
	_v0 := map[int]int{}
	{
		_v1 := 0
	_l0:
		for ; ; _v1++ {
			{
				_v2 := !(_v1 < len(_v0))
				if _v2 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over map (underscore index and underscore value)",
			body: "for _, _ = range map[int]int{} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewMap(intType, intType)}
			},
			expect: `
{
	_v0 := map[int]int{}
	{
		_v1 := 0
	_l0:
		for ; ; _v1++ {
			{
				_v2 := !(_v1 < len(_v0))
				if _v2 {
					break _l0
				}
			}
			foo
		}
	}
}
`,
		},
		{
			name: "for range over map (index only)",
			body: "for i := range map[int]int{} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewMap(intType, intType)}
			},
			expect: `
{
	_v0 := map[int]int{}
	{
		_v1 := make([]int, 0, len(_v0))
		for _v2 := range _v0 {
			_v1 = append(_v1, _v2)
		}
	}
	{
		_v4 := _v1
		{
			_v5 := 0
		_l0:
			for ; ; _v5++ {
				{
					_v6 := !(_v5 < len(_v4))
					if _v6 {
						break _l0
					}
				}
				i := _v4[_v5]
				{
					_, _v3 := _v0[i]
					_v7 := _v3
					if _v7 {
						foo
					}
				}
			}
		}
	}
}
`,
		},
		{
			name: "for range over map (index and underscore value)",
			body: "for i, _ := range map[int]int{} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewMap(intType, intType)}
			},
			expect: `
{
	_v0 := map[int]int{}
	{
		_v1 := make([]int, 0, len(_v0))
		for _v2 := range _v0 {
			_v1 = append(_v1, _v2)
		}
	}
	{
		_v4 := _v1
		{
			_v5 := 0
		_l0:
			for ; ; _v5++ {
				{
					_v6 := !(_v5 < len(_v4))
					if _v6 {
						break _l0
					}
				}
				i := _v4[_v5]
				{
					_, _v3 := _v0[i]
					_v7 := _v3
					if _v7 {
						foo
					}
				}
			}
		}
	}
}
`,
		},
		{
			name: "for range over map (index and value)",
			body: "for i, v := range map[int]int{} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewMap(intType, intType)}
			},
			expect: `
{
	_v0 := map[int]int{}
	{
		_v1 := make([]int, 0, len(_v0))
		for _v2 := range _v0 {
			_v1 = append(_v1, _v2)
		}
	}
	{
		_v4 := _v1
		{
			_v5 := 0
		_l0:
			for ; ; _v5++ {
				{
					_v6 := !(_v5 < len(_v4))
					if _v6 {
						break _l0
					}
				}
				i := _v4[_v5]
				{
					v, _v3 := _v0[i]
					_v7 := _v3
					if _v7 {
						foo
					}
				}
			}
		}
	}
}
`,
		},
		{
			name: "for range over map (underscore index and value)",
			body: "for _, v := range map[int]int{} { foo }",
			info: func(stmts []ast.Stmt, info *types.Info) {
				x := stmts[0].(*ast.RangeStmt).X
				info.Types[x] = types.TypeAndValue{Type: types.NewMap(intType, intType)}
			},
			expect: `
{
	_v0 := map[int]int{}
	{
		_v1 := make([]int, 0, len(_v0))
		for _v2 := range _v0 {
			_v1 = append(_v1, _v2)
		}
	}
	{
		_v5 := _v1
		{
			_v6 := 0
		_l0:
			for ; ; _v6++ {
				{
					_v7 := !(_v6 < len(_v5))
					if _v7 {
						break _l0
					}
				}
				_v3 := _v5[_v6]
				{
					v, _v4 := _v0[_v3]
					_v8 := _v4
					if _v8 {
						foo
					}
				}
			}
		}
	}
}
`,
		},
		{
			name: "select",
			body: `
select {
case <-a:
	foo
case b := <-c:
	bar
case d, ok := <-e:
	baz
case f[g()] = <-h():
	qux
case i() <- j():
	abc
default:
	xyz
}
`,
			types: map[string]types.TypeAndValue{
				"a":  {Type: types.NewChan(types.RecvOnly, intType)},
				"b":  {Type: intType},
				"c":  {Type: types.NewChan(types.RecvOnly, intType)},
				"d":  {Type: intType},
				"e":  {Type: types.NewChan(types.RecvOnly, intType)},
				"ok": {Type: types.Typ[types.Bool]},
			},
			info: func(s []ast.Stmt, info *types.Info) {
				astutil.Apply(s[0], func(cursor *astutil.Cursor) bool {
					ident, ok := cursor.Node().(*ast.Ident)
					if !ok {
						return true
					}
					switch p := cursor.Parent().(type) {
					case *ast.CallExpr:
						switch ident.Name {
						case "h":
							info.Types[p] = types.TypeAndValue{Type: types.NewChan(types.RecvOnly, intType)}
						case "i":
							info.Types[p] = types.TypeAndValue{Type: types.NewChan(types.SendRecv, intType)}
						case "j":
							info.Types[p] = types.TypeAndValue{Type: intType}
						}
					case *ast.IndexExpr:
						switch ident.Name {
						case "f":
							info.Types[p] = types.TypeAndValue{Type: intType}
						}
					}
					return true
				}, nil)
			},
			expect: `
{
	_v0 := 0
	_v1 := a
	_v2 := c
	var _v3 int
	_v4 := e
	var _v5 int
	var _v6 bool
	_v7 := h()
	var _v8 int
	_v9 := i()
	_v10 := j()
	select {
	case <-_v1:
		_v0 = 1
	case _v3 = <-_v2:
		_v0 = 2
	case _v5, _v6 = <-_v4:
		_v0 = 3
	case _v8 = <-_v7:
		_v0 = 4
	case _v9 <- _v10:
		_v0 = 5
	default:
		_v0 = 6
	}
	{
		_v11 := _v0
		switch _v11 {
		case 1:
			foo
		case 2:
			b := _v3
			bar
		case 3:
			d := _v5
			ok := _v6
			baz
		case 4:
			f[g()] = _v8
			qux
		case 5:
			abc
		case 6:
			xyz
		}
	}
}

`,
		},
		{
			name: "select break",
			body: `
label:
	select {
	case <-a:
		break
	case <-b:
		break label
	}
`,
			expect: `
{
	_v0 := 0
	_v1 := a
	_v2 := b
	select {
	case <-_v1:
		_v0 = 1
	case <-_v2:
		_v0 = 2
	}
	{
		_v3 := _v0
	_l0:
		switch _v3 {
		case 1:
			break _l0
		case 2:
			break _l0
		}
	}
}
`,
		},
		{
			name: "switch with init + tag",
			body: `
switch foo := bar; foo {
case 1:
	bar
default:
	baz
}
`,
			expect: `
{
	foo := bar
	_v0 := foo
	switch _v0 {
	case 1:
		bar
	default:
		baz
	}
}
`,
		},
		{
			name: "switch tag",
			body: `
switch foo {
case 1:
	bar
default:
	baz
}
`,
			expect: `
{
	_v0 := foo
	switch _v0 {
	case 1:
		bar
	default:
		baz
	}
}
`,
		},
		{
			name: "raw switch",
			body: `
switch {
case foo == 1:
	bar
case foo == 2:
	baz
default:
	baz
}
`,
			expect: `
switch {
case foo == 1:
	bar
case foo == 2:
	baz
default:
	baz
}
`,
		},
		{
			name: "type switch",
			body: `
switch a.(type) {
case int:
	foo
case bool:
	bar
}
`,
			expect: `
{
	_v0 := a
	switch _v0.(type) {
	case int:
		foo
	case bool:
		bar
	}
}
`,
		},
		{
			name: "type switch with init + assign",
			body: `
switch a := 1; b := a.(type) {
case int:
	foo
case bool:
	bar
}
`,
			expect: `
{
	a := 1
	_v0 := a
	switch b := _v0.(type) {
	case int:
		foo
	case bool:
		bar
	}
}
`,
		},
		{
			name: "nested breaks and continues",
			body: `
l1:
	for {
		break
		continue
	l2:
		select {
		case <-a:
			break
			break l1
			continue
		default:
			l3:
				switch a.(type) {
				case int:
					break
					break l2
					break l1
					continue

					l4:
					switch {
					default:
						break
						break l3
						break l2
						break l1
						continue
					}
				}
			l5:
				for {
					break
					break l2
					break l1
					continue
					continue l1
				}
		}
	}
`,
			defs: map[string]types.Object{
				"l1": types.NewLabel(0, nil, "l1"),
				"l2": types.NewLabel(0, nil, "l2"),
				"l3": types.NewLabel(0, nil, "l3"),
				"l4": types.NewLabel(0, nil, "l4"),
				"l5": types.NewLabel(0, nil, "l5"),
			},
			expect: `
_l0:
	for {
		break _l0
		continue _l0
		{
			_v0 := 0
			_v1 := a
			select {
			case <-_v1:
				_v0 = 1
			default:
				_v0 = 2
			}
			{
				_v2 := _v0
			_l1:
				switch _v2 {
				case 1:
					break _l1
					break _l0
					continue _l0
				case 2:
					{
						_v3 := a
					_l2:
						switch _v3.(type) {
						case int:
							break _l2
							break _l1
							break _l0
							continue _l0
						_l3:
							switch {
							default:
								break _l3
								break _l2
								break _l1
								break _l0
								continue _l0
							}
						}
					}
				_l4:
					for {
						break _l4
						break _l1
						break _l0
						continue _l4
						continue _l0
					}
				}
			}
		}
	}
`,
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
			if test.info != nil {
				test.info(body.List, info)
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
