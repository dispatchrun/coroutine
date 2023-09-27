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
	"golang.org/x/tools/go/packages"
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
			_v1 := i < 10
			_v0 := !_v1
			if _v0 {
				break _l0
			}
		}
		result += i
	}
}
`,
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
			_v1 := i < 10
			_v0 := !_v1
			if _v0 {
				break _l0
			}
		}
		{
			j := 0
		_l1:
			for ; ; j++ {
				{
					_v3 := j < 10
					_v2 := !_v3
					if _v2 {
						break _l1
					}
				}
				break _l0
			}
		}
	}
}
`,
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
		default:
			{
				_v0 := true
				if _v0 {
					break _l1
				} else {
					_v1 := false
					if _v1 {
						continue _l0
					} else {
						break _l0
					}
				}
			}
		}
	}
`,
		},
		{
			name: "for range over slice",
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
		for ; _v1 < len(_v0); _v1++ {
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
		for ; _v1 < len(_v0); _v1++ {
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
		for ; _v1 < len(_v0); _v1++ {
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
		for ; i < len(_v0); i++ {
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
		for ; i < len(_v0); i++ {
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
		for ; i < len(_v0); i++ {
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
		for ; _v1 < len(_v0); _v1++ {
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
		for ; i < len(_v0); i++ {
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
		for ; _v1 < len(_v0); _v1++ {
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
		for ; _v1 < len(_v0); _v1++ {
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
			for ; _v5 < len(_v4); _v5++ {
				i := _v4[_v5]
				{
					_, _v3 := _v0[i]
					if _v3 {
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
			for ; _v5 < len(_v4); _v5++ {
				i := _v4[_v5]
				{
					_, _v3 := _v0[i]
					if _v3 {
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
			for ; _v5 < len(_v4); _v5++ {
				i := _v4[_v5]
				{
					v, _v3 := _v0[i]
					if _v3 {
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
			for ; _v6 < len(_v5); _v6++ {
				_v3 := _v5[_v6]
				{
					v, _v4 := _v0[_v3]
					if _v4 {
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
		switch {
		default:
			{
				_v12 := _v11 == 1
				if _v12 {
					foo
				} else {
					_v13 := _v11 == 2
					if _v13 {
						b := _v3
						bar
					} else {
						_v14 := _v11 == 3
						if _v14 {
							d := _v5
							ok := _v6
							baz
						} else {
							_v15 := _v11 == 4
							if _v15 {
								_v18 := g()
								f[_v18] = _v8
								qux
							} else {
								_v16 := _v11 == 5
								if _v16 {
									abc
								} else {
									_v17 := _v11 == 6
									if _v17 {
										xyz
									}
								}
							}
						}
					}
				}
			}
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
		switch {
		default:
			{
				_v4 := _v3 == 1
				if _v4 {
					break _l0
				} else {
					_v5 := _v3 == 2
					if _v5 {
						break _l0
					}
				}
			}
		}
	}
}
`,
		},
		{
			name: "empty select",
			body: `
select {}
`,
			expect: `
select {}
`,
		},
		{
			name: "empty select with default",
			body: `
select {
default:
}
`,
			expect: `
{
	_v0 := 0
	select {
	default:
		_v0 = 1
	}
	switch _v0 {
	case 1:
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
	switch {
	default:
		{
			_v1 := _v0 == 1
			if _v1 {
				bar
			} else {
				baz
			}
		}
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
case 2, 3, 4:
	baz
default:
	qux
}
`,
			expect: `
{
	_v0 := foo
	switch {
	default:
		{
			_v1 := _v0 == 1
			if _v1 {
				bar
			} else {
				_v2 := (_v0 == 2) | (_v0 == 3) | (_v0 == 4)
				if _v2 {
					baz
				} else {
					qux
				}
			}
		}
	}
}
`,
		},
		{
			name: "empty switch",
			body: "switch {}",
			// TODO: remove the unnecessary default case
			expect: `
switch {
default:
}`,
		},
		{
			name: "empty switch with default",
			body: "switch { default: }",
			// TODO: remove empty stmt inside default case
			expect: `
switch {
default:
	{
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
default:
	qux
case foo == 2:
	baz
}`,
			expect: `
switch {
default:
	{
		_v0 := foo == 1
		if _v0 {
			bar
		} else {
			_v1 := foo == 2
			if _v1 {
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
			name: "type switch",
			body: `
switch a.(type) {
case int:
	foo
case bool, string:
	bar
}
`,
			expect: `
switch a.(type) {
case int:
	foo
case bool, string:
	bar
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
	switch b := a.(type) {
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
				switch {
				default:
					{
						_v3 := _v2 == 1
						if _v3 {
							break _l1
							break _l0
							continue _l0
						} else {
							_v4 := _v2 == 2
							if _v4 {
								{
								_l2:
									switch a.(type) {
									case int:
										break _l2
										break _l1
										break _l0
										continue _l0
									_l3:
										switch {
										default:
											{
												break _l3
												break _l2
												break _l1
												break _l0
												continue _l0
											}
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
			}
		}
	}
`,
		},
		{
			name: "decompose expressions in expr statements",
			body: "a(b(c(d(e(1 + 2)))))",
			expect: `
{
	_v4 := 1 + 2
	_v3 := e(_v4)
	_v2 := d(_v3)
	_v1 := c(_v2)
	_v0 := b(_v1)
	a(_v0)
}
`,
		},
		{
			name: "decompose expressions in incdec statements",
			body: "a(b())++",
			expect: `
{
	_v0 := b()
	a(_v0)++
}
`,
		},
		{
			name: "decompose expressions in decl statements",
			body: "var _, _ int = a(b(0)), c(d(1))",
			// See https://go.dev/play/p/PkwoJbDLgQV for order of evaluation.
			expect: `
{
	_v1 := b(0)
	_v0 := a(_v1)
	_v3 := d(1)
	_v2 := c(_v3)
	var _, _ int = _v0, _v2
}
`,
		},
		{
			name: "decompose expressions in assignment statements",
			body: "ints[a(b(0))], ints[c(d(1))] = e(f(10)), g(h(11))",
			// See https://go.dev/play/p/WvrxhauFbsA for order of evaluation
			expect: `
{
	_v1 := b(0)
	_v0 := a(_v1)
	_v3 := d(1)
	_v2 := c(_v3)
	_v5 := f(10)
	_v4 := e(_v5)
	_v7 := h(11)
	_v6 := g(_v7)
	ints[_v0], ints[_v2] = _v4, _v6
}
`,
		},
		{
			name: "decompose expressions in return statements",
			body: "return a(b(0)), c(d(1))",
			// See https://go.dev/play/p/PkwoJbDLgQV for order of evaluation.
			expect: `
{
	_v1 := b(0)
	_v0 := a(_v1)
	_v3 := d(1)
	_v2 := c(_v3)
	return _v0, _v2
}
`,
		},
		{
			name: "decompose expressions in send statements",
			body: "a(b()) <- c(d())",
			expect: `
{
	_v1 := b()
	_v0 := a(_v1)
	_v3 := d()
	_v2 := c(_v3)
	_v0 <- _v2
}
`,
		},
		{
			name: "FIXME: don't hoist function calls when there are multiple results",
			body: "a, b, c = d(e())",
			info: func(stmts []ast.Stmt, info *types.Info) {
				callExpr := stmts[0].(*ast.AssignStmt).Rhs[0]
				info.Types[callExpr] = types.TypeAndValue{
					Type: types.NewTuple(
						types.NewVar(0, nil, "a", intType),
						types.NewVar(0, nil, "b", intType),
						types.NewVar(0, nil, "c", intType),
					),
				}
			},
			expect: `
{
	_v0 := e()
	a, b, c = d(_v0)
}
`,
		},
		{
			name: "key value expr",
			body: "Foo{Bar: a(b()), Baz: c(d())}",
			// TODO: fix order of evaluation here
			expect: `
{
	_v3 := d()
	_v2 := b()
	_v1 := c(_v3)
	_v0 := a(_v2)
	Foo{Bar: _v0, Baz: _v1}
}
`,
		},
		{
			name: "defer with func literal",
			body: "defer func() { foo() }()",
			expect: `
defer func() {
	foo()
}()
`,
		},
		{
			name: "defer with func literal args",
			body: "defer func() { foo() }(a, b, c)",
			expect: `
{
	_v0 := a
	_v1 := b
	_v2 := c
	defer func() {
		func() {
			foo()
		}(_v0, _v1, _v2)
	}()
}
`,
		},
		{
			name: "defer with func literal and internal args",
			body: "defer func() { foo(a, b, c) }()",
			expect: `
defer func() {
	foo(a, b, c)
}()
`,
		},
		{
			name: "defer without func literal",
			body: "defer foo()",
			expect: `
defer func() {
	foo()
}()
`,
		},
		{
			name: "defer without func literal args",
			body: "defer foo(a(b()), c)",
			expect: `
{
	_v2 := b()
	_v0 := a(_v2)
	_v1 := c
	defer func() {
		foo(_v0, _v1)
	}()
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
			// We're testing worst case desugaring, so mark all nodes
			// as possibly yielding.
			mayYield := map[ast.Node]struct{}{}

			ast.Inspect(body, func(node ast.Node) bool {
				if node != nil {
					mayYield[node] = struct{}{}
				}
				if ident, ok := node.(*ast.Ident); ok {
					if obj, ok := test.defs[ident.Name]; ok {
						info.Defs[ident] = obj
					} else if obj, ok := test.uses[ident.Name]; ok {
						info.Uses[ident] = obj
					} else {
						// Unless an override has been specified, link
						// identifiers to objects defined in types.Universe.
						if obj := types.Universe.Lookup(ident.Name); obj != nil {
							info.Uses[ident] = obj
						}
					}
					if t, ok := test.types[ident.Name]; ok {
						info.Types[ident] = t
					}
				}
				return true
			})

			p := &packages.Package{TypesInfo: info}
			desugared := desugar(p, body, mayYield, nil, nil)
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
	// ast.Print(fset, node)
	var b bytes.Buffer
	if err := format.Node(&b, fset, node); err != nil {
		panic(err)
	}
	return b.String()
}
