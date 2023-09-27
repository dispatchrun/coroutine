//go:build !durable

package main

import (
	"errors"
	"log"
	"os"

	"github.com/stealthrocket/coroutine"
)

func main() {
	scraper := NewScraper("https://en.wikipedia.org/wiki/Main_Page", 100)

	coro := coroutine.New[string, any](func() {
		scraper.Start()
	})

	// Restore state.
	if state, err := os.ReadFile(".state"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	} else if _, err := coro.Context().Unmarshal(state); err != nil {
		log.Fatal(err)
	}

	for coro.Next() {
		url := coro.Recv()
		log.Println("scraped", url)

		// Persist state.
		if b, err := coro.Context().Marshal(); err != nil {
			log.Fatal(err)
		} else if err := os.WriteFile(".state", b, 0644); err != nil {
			log.Fatal(err)
		}
	}
}
