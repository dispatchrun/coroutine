//go:build !durable

package main

import (
	"errors"
	"log"
	"os"

	"github.com/stealthrocket/coroutine"
)

func main() {
	s := NewScraper("https://en.wikipedia.org/wiki/Main_Page", 100)

	c := coroutine.New[string, any](func() {
		s.Start()
	})

	// Restore state.
	ctx := c.Context()
	state, err := os.ReadFile(".state")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	} else if _, err := ctx.Unmarshal(state); err != nil {
		log.Fatal(err)
	}

	for c.Next() {
		url := c.Recv()
		log.Println("scraped", url)

		b, err := c.Context().Marshal()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(".state", b, 0644); err != nil {
			log.Fatal(err)
		}
	}
}
