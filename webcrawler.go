package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Has url been fetched? Respond on the channel `respond`.
type Fetched struct {
        url string
        respond chan bool
}

type Result struct {
        url string
        body string
        err error
}

// crawlImpl uses fetcher to recursively crawl pages starting with url, to a
// maximum of depth. It signals that it's finished by sending a message on
// done, checks if the URL it's crawling has been done already by querying
// fetched, and sends the result of fetching its url to output.
func crawlImpl(url string, depth int, fetcher Fetcher, done chan bool,
                fetched chan<- Fetched, output chan Result) {
        response := make(chan bool)
        fetched <- Fetched{ url, response }
	if <-response || depth <= 0 {
		done <- true
		return
	}
	body, urls, err := fetcher.Fetch(url)
        output <- Result{ url, body, err }
	if err != nil {
		fmt.Println(err)
	} else {
                children := make(chan bool)
                for _, u := range urls {
                        go crawlImpl(u, depth-1, fetcher, children, fetched, output)
                }
                for i := 0; i < len(urls); i++ {
                        <-children
                }
                close(children)
        }
        done <- true
	return
}

// Crawl uses fetcher to recursively crawl pages starting with url, to a
// maximum of depth, and sends the result of fetching its url to output.
func Crawl(url string, depth int, fetcher Fetcher, output chan Result) {
        fetched := make(chan Fetched)
        go func() {
                seen := make(map[string]bool)
                for query := range fetched {
                        query.respond <- seen[query.url]
                        seen[query.url] = true
                }
        }()
        done := make(chan bool)
        go func() {
                <-done
        }()
	crawlImpl(url, depth, fetcher, done, fetched, output)
	close(fetched)
        close(output)
}

func main() {
	ch := make(chan Result)
	go Crawl("http://golang.org/", 4, fetcher, ch)
	for result := range ch {
                if result.err == nil {
                        fmt.Println(result.url, ":", result.body)
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
