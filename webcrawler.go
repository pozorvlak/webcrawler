package main

import (
    "fmt"
)

type Fetcher interface {
    // Fetch returns the body of URL and
    // a slice of URLs found on that page.
    Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func crawlImpl(url string, depth int, fetcher Fetcher, ch chan []string) {
    // TODO: Fetch URLs in parallel.
    // TODO: Don't fetch the same URL twice.
    // This implementation doesn't do either:
	crawled := make([]string, 0)
    if depth <= 0 {
		ch <- crawled
        return
    }
    _, urls, err := fetcher.Fetch(url)
    if err != nil {
        fmt.Println(err)
		ch <- crawled
        return
    }
	crawled = []string{ url }
	subCh := make(chan []string)
    for _, u := range urls {
        go crawlImpl(u, depth-1, fetcher, subCh)
    }
	for i := 0; i < len(urls); i++ {
		us := <-subCh
		crawled = append(crawled, us...)
    }
	close(subCh)
	ch <- crawled
    return
}

func Crawl(url string, depth int, fetcher Fetcher, ch chan []string) {
	crawlImpl(url, depth, fetcher, ch)
	close(ch)
}

func main() {
    ch := make(chan []string)
    go Crawl("http://golang.org/", 4, fetcher, ch)
	for u := range ch {
		fmt.Println("Received at toplevel: ", u)
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
