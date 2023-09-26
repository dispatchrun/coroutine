//go:build !durable

package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/stealthrocket/coroutine"
)

// Scraper recursively scrapes URLs.
type Scraper struct {
	url   string
	limit int
	seen  map[string]struct{}
}

// NewScraper creates a new Scraper that starts at the
// specified URL and scrapes up to limit unique URLs
// from the same host.
func NewScraper(url string, limit int) *Scraper {
	return &Scraper{
		url:   url,
		limit: limit,
		seen:  map[string]struct{}{url: {}},
	}
}

func (s *Scraper) Start() {
	queue := []string{s.url}
	for i := 0; i < len(queue); i++ {
		links, err := s.scrape(queue[i])
		if err != nil {
			log.Printf("warning: %s => %v", queue[i], err)
			continue
		}
		for _, link := range links {
			if _, ok := s.seen[link]; !ok {
				queue = append(queue, link)
				s.seen[link] = struct{}{}
			}
		}
	}
}

func (s *Scraper) scrape(url string) ([]string, error) {
	res, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	coroutine.Yield[string, any](url)

	return collectLinks(url, body), nil
}

// This is just a quick demo to highlight coroutines rather
// than HTML parsing. Don't parse HTML with regexp!
var ahrefs = regexp.MustCompile(`<a[^>]* href=['"]([^'"]+)['"]`)

func collectLinks(parentURL string, body []byte) []string {
	parent, _ := url.Parse(parentURL)

	var links []string
	for _, match := range ahrefs.FindAllSubmatch(body, -1) {
		href, err := url.Parse(string(match[1]))
		if err != nil {
			// Skip invalid/unsupported URLs.
			continue
		}
		if href.Host != "" && href.Host != parent.Host {
			// Only scrape pages from the same host.
			continue
		}
		// Fill in missing host/scheme.
		href.Host = parent.Host
		href.Scheme = parent.Scheme
		// Keep it simple. Trim query+fragment and only scrape unique paths.
		href.RawPath, _, _ = strings.Cut(href.RawPath, "#")
		href.RawPath, _, _ = strings.Cut(href.RawPath, "?")
		href.RawQuery = ""
		href.RawFragment = ""

		links = append(links, href.String())
	}

	return links
}
