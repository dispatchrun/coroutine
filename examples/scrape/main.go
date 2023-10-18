//go:build !durable

package main

import (
	"errors"
	"log"
	"os"

	"github.com/stealthrocket/coroutine"
)

func main() {
	const url = "https://en.wikipedia.org/wiki/Main_Page"
	const limit = 100
	const state = "coroutine.state"

	scraper := NewScraper(url, limit)

	coro := coroutine.New[string, any](func() {
		scraper.Start()
	})

	// Restore state.
	if coroutine.Durable {
		if state, err := os.ReadFile(state); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatal(err)
			}
		} else if _, err := coro.Context().Unmarshal(state); err != nil {
			if errors.Is(err, coroutine.ErrInvalidState) {
				log.Println("warning: coroutine state is no longer valid. Starting fresh")
			} else {
				log.Fatal(err)
			}
		}
	}

	for coro.Next() {
		url := coro.Recv()
		log.Println("scraped", url)

		// Persist state.
		if coroutine.Durable {
			if b, err := coro.Context().Marshal(); err != nil {
				log.Fatal(err)
			} else if err := os.WriteFile(state, b, 0644); err != nil {
				log.Fatal(err)
			}
		}
	}
}
