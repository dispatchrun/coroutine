//go:build !durable

package main

import (
	"fmt"
	"net/http"

	"github.com/dispatchrun/coroutine"
)

type yieldingRoundTripper struct{}

func (*yieldingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	res := coroutine.Yield[*http.Request, *http.Response](req)
	return res, nil
}

func work() {
	res, err := http.Get("http://example.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(res.StatusCode)
}

func main() {
	http.DefaultTransport = &yieldingRoundTripper{}

	c := coroutine.New[*http.Request, *http.Response](work)

	for c.Next() {
		req := c.Recv()
		fmt.Println("Requesting", req.URL.String())
		if req.URL.String() == "http://example.com" {
			c.Send(&http.Response{
				StatusCode: 301,
				Header: http.Header{
					"Location": []string{"https://example.com"},
				},
			})
		} else {
			c.Send(&http.Response{StatusCode: 200})
		}
	}
}
