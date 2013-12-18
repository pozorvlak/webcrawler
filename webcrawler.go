package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Done struct {
        url string
        respond chan bool
}

type Result struct {
        body string
        err error
}

type ResultChan chan map[string]Result

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func crawlImpl(url string, depth int, fetcher Fetcher, ch ResultChan, done chan<- Done) {
	crawled := make(map[string]Result)
        response := make(chan bool)
        done <- Done{ url, response }
	if <-response || depth <= 0 {
		ch <- crawled
		return
	}
	body, urls, err := fetcher.Fetch(url)
        crawled[url] = Result{ body, err }
	if err != nil {
		fmt.Println(err)
		ch <- crawled
		return
	}
	subCh := make(ResultChan)
	for _, u := range urls {
		go crawlImpl(u, depth-1, fetcher, subCh, done)
	}
	for i := 0; i < len(urls); i++ {
		us := <-subCh
                for k, r := range us {
                        crawled[k] = r
                }
	}
	close(subCh)
	ch <- crawled
	return
}

func Crawl(url string, depth int, fetcher Fetcher, ch ResultChan) {
        done := make(chan Done)
        go func() {
                seen := make(map[string]bool)
                for query := range done {
                        query.respond <- seen[query.url]
                        seen[query.url] = true
                }
        }()
	crawlImpl(url, depth, fetcher, ch, done)
	close(ch)
}

func main() {
	ch := make(ResultChan)
	go Crawl("http://golang.org/", 4, fetcher, ch)
        urls := <-ch
	for url, result := range urls {
                if result.err == nil {
                        fmt.Println(url, ":", result.body)
                }
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
