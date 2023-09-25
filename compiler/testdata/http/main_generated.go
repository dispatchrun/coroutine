//go:build durable

package main

import (
	fmt "fmt"
	http "net/http"
	coroutine "github.com/stealthrocket/coroutine"
)
import _types "github.com/stealthrocket/coroutine/types"

type yieldingRoundTripper struct{}
//go:noinline
func (*yieldingRoundTripper) RoundTrip(req *http.Request) (_ *http.Response, _ error) {
	_c := coroutine.LoadContext[*http.Request, *http.Response]()
	_f, _fp := _c.Push()
	var _f0 *struct {
		X0 *http.Request
		X1 *http.Response
	}
	if _f.IP == 0 {
		_f0 = &struct {
			X0 *http.Request
			X1 *http.Response
		}{X0: req}
	} else {
		_f0 = _f.Get(0).(*struct {
			X0 *http.Request
			X1 *http.Response
		})
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _f0)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_f0.X1 = coroutine.Yield[*http.Request, *http.Response](_f0.X0)
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		return _f0.X1, nil
	}
	return
}
//go:noinline
func work() {
	_c := coroutine.LoadContext[*http.Request, *http.Response]()
	_f, _fp := _c.Push()
	var _f0 *struct {
		X0 *http.Response
		X1 error
	}
	if _f.IP == 0 {
		_f0 = &struct {
			X0 *http.Response
			X1 error
		}{}
	} else {
		_f0 = _f.Get(0).(*struct {
			X0 *http.Response
			X1 error
		})
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _f0)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_f0.X0, _f0.X1 = http.Get("http://example.com")
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		if _f0.X1 != nil {
			panic(_f0.X1)
		}
		_f.IP = 3
		fallthrough
	case _f.IP < 4:
		fmt.Println(_f0.X0.StatusCode)
	}
}

func main() {
	http.DefaultTransport = &yieldingRoundTripper{}

	c := coroutine.New[*http.Request, *http.Response](work)

	for c.Next() {
		req := c.Recv()
		fmt.Println("Requesting", req.URL.String())
		c.Send(&http.Response{
			StatusCode: 200,
		})
	}
}
func init() {
	_types.RegisterFunc[func(req *http.Request) (_ *http.Response, _ error)]("github.com/stealthrocket/coroutine/compiler/testdata/http.RoundTrip")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata/http.main")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata/http.work")
}
